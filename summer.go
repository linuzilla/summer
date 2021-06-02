package summer

import (
	"container/list"
	"fmt"
	"github.com/linuzilla/summer/gobean"
	"github.com/linuzilla/summer/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
)

const DefaultInjectionTag = `inject`
const DefaultPluginNamePrefix = `plugin#`

type contextManagerImpl struct {
	items                    *list.List
	itemsMap                 map[string]*gobean.PopulateItem
	debug                    bool
	injectionTag             string
	pluginNamePrefix         string
	setterNameFunc           func(variableName string) string
	exportedVariableNameFunc func(variableName string) string
}

func (ctx *contextManagerImpl) addBean(bean interface{}) (*gobean.PopulateItem, error) {
	if item, err := gobean.New(bean, 3, ctx.injectionTag); err != nil {
		return nil, err
	} else {
		ctx.items.PushBack(item)
		return item, nil
	}
}

func (ctx *contextManagerImpl) Add(beans ...interface{}) ApplicationContextManager {
	for _, bean := range beans {
		if _, err := ctx.addBean(bean); err != nil {
			panic(err)
		}
	}
	return ctx
}

func (ctx *contextManagerImpl) AddWithName(beanName string, bean interface{}) ApplicationContextManager {
	if _, found, _ := ctx.getBeanByName(beanName); !found {
		if item, err := ctx.addBean(bean); err != nil {
			panic(err)
		} else {
			ctx.itemsMap[beanName] = item
		}
	} else {
		panic(fmt.Errorf("duplicate bean name:'%s'", beanName))
	}
	return ctx
}

func (ctx *contextManagerImpl) assignable(item *gobean.PopulateItem, modelType reflect.Type) bool {
	switch modelType.Kind() {
	case reflect.Interface:
		if item.BeanType.Implements(modelType) {
			return true
		}

	case reflect.Struct:
		if item.BeanType.Elem() == modelType {
			return true
		}
	}
	return false
}

func (ctx *contextManagerImpl) findByType(modelType reflect.Type, singleMatchOnly bool, callback func(item *gobean.PopulateItem)) (*gobean.PopulateItem, int) {
	var rc *gobean.PopulateItem = nil
	var candidate *gobean.PopulateItem = nil
	matched := 0

	for e := ctx.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*gobean.PopulateItem)

		duplicate := false

		switch modelType.Kind() {
		case reflect.Interface:
			if item.BeanType.Implements(modelType) {
				matched++
				if item.Wired {
					if callback != nil {
						callback(item)
					}
					candidate = item
				}

				if rc == nil {
					rc = item
				} else {
					duplicate = true
				}
			}

		case reflect.Struct:
			if item.BeanType.Elem() == modelType {
				matched++
				if item.Wired {
					if callback != nil {
						callback(item)
					}
					candidate = item
				}
				if rc == nil {
					rc = item
				} else {
					duplicate = true
				}
			}
		}

		if duplicate {
			if singleMatchOnly {
				if matched == 2 {
					fmt.Println("Multiple match for [", modelType, "]:")
					fmt.Println(">> ", rc.BeanType)
				}
				fmt.Println(">> ", item.BeanType)
				fmt.Println("Consider using match by name instead.")
			}
		}

	}

	return candidate, matched
}

func (ctx *contextManagerImpl) Get(expectedTypeData interface{}) (interface{}, error) {
	if item, matched := ctx.findByType(reflect.TypeOf(expectedTypeData).Elem(), true, nil); item != nil {
		if matched > 1 {
			return nil, fmt.Errorf("multiple match found")
		} else {
			if reflect.TypeOf(expectedTypeData).Kind() == reflect.Ptr {
				if elem := reflect.ValueOf(expectedTypeData).Elem(); elem.CanSet() {
					elem.Set(item.BeanValue)
				}
			}
			return item.Bean, nil
		}
	} else {
		return nil, fmt.Errorf("no match found")
	}
}

func (ctx *contextManagerImpl) ForEach(match interface{}, callback func(data interface{})) int {
	rc := 0
	ctx.findByType(reflect.TypeOf(match).Elem(), false, func(item *gobean.PopulateItem) {
		callback(item.Bean)
		rc++
	})
	return rc
}

func (ctx *contextManagerImpl) Each(callback func(data interface{})) int {
	rc := 0
	for e := ctx.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*gobean.PopulateItem)

		if item.Wired && callback != nil {
			callback(item.Bean)
			rc++
		}
	}
	return rc
}

func (ctx *contextManagerImpl) getBeanByName(beanName string) (*gobean.PopulateItem, bool, error) {
	if item, found := ctx.itemsMap[beanName]; found {
		if item.Wired {
			return item, found, nil
		} else {
			return nil, found, fmt.Errorf("bean name '%s' not ready fully wired yet", beanName)
		}
	} else {
		return nil, found, fmt.Errorf("bean name '%s' not found", beanName)
	}
}

func (ctx *contextManagerImpl) GetByName(beanName string) (interface{}, error) {
	if item, found, err := ctx.getBeanByName(beanName); !found {
		return nil, err
	} else if item != nil {
		return item.Bean, nil
	} else {
		return nil, fmt.Errorf("bean name '%s' not found", beanName)
	}
}

func (ctx *contextManagerImpl) findWiredEntryByType(modelType reflect.Type) (matchedItem *gobean.PopulateItem, matchCount int) {
	matchedItem = nil
	matchCount = 0

	for e := ctx.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*gobean.PopulateItem)

		switch modelType.Kind() {
		case reflect.Interface:
			if item.BeanType.Implements(modelType) {
				matchCount++

				if item.Wired && matchedItem == nil { // match first
					matchedItem = item
				}
			}

		case reflect.Struct:
			if item.BeanType.Elem() == modelType {
				matchCount++
				if item.Wired && matchedItem == nil { // match first
					matchedItem = item
				}
			}
		}
	}

	return matchedItem, matchCount
}

func (ctx *contextManagerImpl) setValueToField(item *gobean.PopulateItem, elemField *gobean.ElementField, matchedItem *gobean.PopulateItem) error {
	field := elemField.FieldValue
	elemFieldStruct := elemField.StructField

	setterMethodName := ctx.setterNameFunc(elemFieldStruct.Name)

	setter := item.BeanValue.MethodByName(setterMethodName)

	if setter.IsValid() {
		if elemFieldStruct.Type.Kind() == reflect.Ptr {
			if elemFieldStruct.Type == matchedItem.BeanType {
				setter.Interface().(func(interface{}))(matchedItem.Bean)
			} else {
				log.Fatal("Autowire", elemFieldStruct.Type, "via", setterMethodName, "on", item.BeanType)
			}
		} else {
			setter.Interface().(func(interface{}))(matchedItem.Bean)
		}
	} else if field.CanSet() {
		field.Set(matchedItem.BeanValue)
	} else {
		return fmt.Errorf("%s: No setter (%s) or structField not settable", elemFieldStruct.Type, setterMethodName)
	}

	return nil
}

func (ctx *contextManagerImpl) injectMatchedBean(item *gobean.PopulateItem, elemField *gobean.ElementField, matchedItem *gobean.PopulateItem, byName bool) error {
	if err := ctx.setValueToField(item, elemField, matchedItem); err != nil {
		fmt.Println(err)
		return err
	} else if !elemField.Wired {
		elemField.Wired = true
		item.WiredCount++
		if item.CheckIsWired() {
			if postConstructable, ok := item.Bean.(HavePostConstruct); ok {
				if ctx.debug {
					fmt.Printf("PostConstruct: %s\n", item.BeanType.String())
				}
				postConstructable.PostSummerConstruct()
			}
		}

		if ctx.debug {
			if byName {
				fmt.Printf("Auto wire for [%s] : [%s] by name: [%s]\n", item.BeanType.String(), elemField.StructField.Type.String(), elemField.TagValue)
			} else {
				fmt.Printf("Auto wire for [%s] : [%s]\n", item.BeanType.String(), elemField.StructField.Type.String())
			}
		}
	}
	return nil
}

func (ctx *contextManagerImpl) injectField(item *gobean.PopulateItem, elemField *gobean.ElementField) (bool, error) {
	haveInjection := false

	switch {
	case elemField.TagValue == `*`: // injectMatchedBean by type

		elemFieldType := elemField.StructField.Type

		if elemField.StructField.Type.Kind() == reflect.Interface {
		} else {
			elemFieldType = elemField.StructField.Type.Elem()
		}

		matchedItem, cnt := ctx.findWiredEntryByType(elemFieldType)

		switch {
		case cnt == 1 && matchedItem != nil:
			if err := ctx.injectMatchedBean(item, elemField, matchedItem, false); err != nil {
				return false, err
			}
			haveInjection = true

		case cnt > 1:
			fmt.Printf("Number of matched item: %d (structField index=%d)\n", cnt, elemField.Index)

		case cnt == 0:
			return false, fmt.Errorf("%s: no suitable bean\n", elemField.FullName(ctx.injectionTag))
		}

	default: // injectMatchedBean by name
		if matchedItem, found, err := ctx.getBeanByName(elemField.TagValue); matchedItem != nil {
			if err := ctx.injectMatchedBean(item, elemField, matchedItem, true); err != nil {
				return haveInjection, err
			}
			haveInjection = true
		} else if found {
		} else {
			ctx.dumpPendingInjectionField(item)
			return haveInjection, fmt.Errorf("%s: %s", elemField.FullName(ctx.injectionTag), err)
		}

	}
	return haveInjection, nil
}

func (ctx *contextManagerImpl) dumpPendingInjectionField(item *gobean.PopulateItem) {
	fmt.Printf(">> %s\n", item.Source)
	for _, elemField := range item.Fields {
		if !elemField.Wired {
			fmt.Printf(".... %s\n", elemField.FullName(ctx.injectionTag))
		}
	}
}

func (ctx *contextManagerImpl) dumpPendingInjection() {
	for e := ctx.items.Front(); e != nil; e = e.Next() {
		if item, ok := e.Value.(*gobean.PopulateItem); ok {
			if !item.Wired {
				ctx.dumpPendingInjectionField(item)
			}
		}
	}
}

func (ctx *contextManagerImpl) performDependencyInjection() error {
	for i := 1; true; i++ {
		done := true
		makeProgress := false

		for e := ctx.items.Front(); e != nil; e = e.Next() {
			if item, ok := e.Value.(*gobean.PopulateItem); ok {
				if !item.Wired {
					done = false

					for _, elemField := range item.Fields {
						if !elemField.Wired {
							if haveInject, err := ctx.injectField(item, elemField); err != nil {
								return err
							} else if haveInject {
								makeProgress = true
							}
						}
					}
				}
			}
		}

		if done {
			return nil
		} else if !makeProgress {
			ctx.dumpPendingInjection()
			return fmt.Errorf("failed to autowiring")
		}
	}
	return nil
}

func (ctx *contextManagerImpl) Autowiring(callback func(err error)) chan error {
	errorChannel := make(chan error)

	go func() {
		err := ctx.performDependencyInjection()
		callback(err)
		errorChannel <- err
	}()
	return errorChannel
}

func (ctx *contextManagerImpl) PerformAutoWiring(onError func(err error)) ApplicationContextManager {
	if err := ctx.performDependencyInjection(); err != nil {
		if onError != nil {
			onError(err)
		} else {
			log.Fatal(err)
		}
	}
	return ctx
}

func (ctx *contextManagerImpl) LoadPlugins(path string, callback func(beanName string, file string, module interface{}, err error)) error {
	if files, err := ioutil.ReadDir(path); err != nil {
		return err
	} else {
		Stream(files).Filter(func(i interface{}) bool {
			return strings.HasSuffix(i.(os.FileInfo).Name(), `.so`)
		}).Map(func(i interface{}) interface{} {
			return i.(os.FileInfo).Name()
		}).ForEach(func(i interface{}) {
			moduleName := i.(string)
			exportedVariableName := ctx.exportedVariableNameFunc(moduleName)
			beanName := ctx.pluginNamePrefix + exportedVariableName

			if plug, err := plugin.Open(filepath.Join(path, moduleName)); err != nil {
				if callback != nil {
					callback(beanName, moduleName, nil, err)
				}
			} else {
				module, err := plug.Lookup(exportedVariableName)

				if err == nil {
					ctx.AddWithName(beanName, module)
				}

				if callback != nil {
					callback(beanName, moduleName, module, err)
				}
			}
		})
	}
	return nil
}

// Setters

func (ctx *contextManagerImpl) SetExportedVariableNameFunc(function func(string) string) {
	ctx.exportedVariableNameFunc = function
}

func (ctx *contextManagerImpl) SetSetterNameFunc(function func(string) string) {
	ctx.setterNameFunc = function
}

func (ctx *contextManagerImpl) SetTagName(tagName string) {
	ctx.injectionTag = tagName
}

func (ctx *contextManagerImpl) SetPluginBeanNamePrefix(prefix string) {
	ctx.pluginNamePrefix = prefix
}

func (ctx *contextManagerImpl) Debug(on bool) {
	ctx.debug = on
}

func New() ApplicationContextManager {
	return &contextManagerImpl{
		items:                    list.New(),
		itemsMap:                 map[string]*gobean.PopulateItem{},
		injectionTag:             DefaultInjectionTag,
		pluginNamePrefix:         DefaultPluginNamePrefix,
		debug:                    false,
		setterNameFunc:           utils.SetterName,
		exportedVariableNameFunc: utils.FileNameToExportedVariable,
	}
}

func Initialize(callback func(ApplicationContextManager), onError func(error)) ApplicationContextManager {
	ctx := New()
	{
		defer func() {
			if e := recover(); e != nil {
				var err error

				if anError, ok := e.(error); ok {
					err = anError
				} else {
					err = fmt.Errorf("%v", e)
				}

				if onError != nil {
					onError(err)
				} else {
					log.Fatal(err)
				}
			}
		}()
		callback(ctx)
	}
	return ctx
}
