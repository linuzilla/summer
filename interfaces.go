// Try to provide "dependency injection" mechanism on the Go world.
package summer

// kind of like "@PostConstruct" in Spring framework
type HavePostConstruct interface {
	PostSummerConstruct()
}

type ApplicationContextManager interface {
	// Add "beans" to, the "bean" should be a "pointer" or "interface", however, "pointer to interface" is not recommended.
	Add(beans ...interface{}) ApplicationContextManager

	// To avoid more the one candidate "beans", use name to distinguish between them.
	AddWithName(beanName string, bean interface{}) ApplicationContextManager

	// To perform dependency injection.
	Autowiring(callback func(err error)) chan error

	// a newer version of "Autowiring" function
	PerformAutoWiring(onError func(err error)) ApplicationContextManager

	// retrieve bean based on argument variable type, argument should be a "pinter to interface" or "pointer to structure".
	Get(intf interface{}) (interface{}, error)

	// retrieve bean by name
	GetByName(beanName string) (interface{}, error)

	// iterate over every beans which match the interface in the context
	ForEach(match interface{}, callback func(data interface{})) int

	// iterate over all wired beans
	Each(callback func(data interface{})) int

	// plugin can be loaded and put them in the context and perform dependency injection as well.
	// every plugin can only inject a variable using plugin's filename as variable name
	// LoadPlugins will try the find the exported variable using "FileNameToExportedVariable" function
	LoadPlugins(path string, callback func(beanName string, file string, module interface{}, err error)) error

	// If the default "exported variable name" converting function not suite for you,
	// provide a relevant one.
	SetExportedVariableNameFunc(function func(string) string)

	// To distinguish between plugins loaded by LoadPlugins, every plugin will added to application context
	// with a name associated with it.  The name, by default, will be the "variable name" prefixed by "plugin#",
	// However, you can choose the prefix you wanted calling "SetPluginsBeanNamePrefix" before "LoadPlugins"
	SetPluginBeanNamePrefix(prefix string)

	// By default, The setter name of a variable is follow Java's setter idea with the first letter 'S' capitalized.
	// However, there is no standard "setter" function in Go world
	SetSetterNameFunc(function func(string)string)

	// The field wanted to be inject require a tag, the default tag is 'inject'.
	// Change tag name if other name is desired
	SetTagName(tagName string)

	Debug(on bool)
}
