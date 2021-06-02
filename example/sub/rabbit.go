package sub

import "fmt"

type Rabbit struct {
	// Husky *Dog `inject:"*"`
}

func (r *Rabbit) Jump() {
	fmt.Println("Rabbit jump")
}
