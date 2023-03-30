package core

// NOT thread-safe
type Debounced struct {
	state bool
}

func (d *Debounced) Invoke() bool {
	if !d.state {
		d.state = true
		return true
	} else {
		return false
	}
}

func (d *Debounced) Reset() {
	d.state = false
}
