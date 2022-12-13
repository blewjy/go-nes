package nes

// SetC will set the C flag if the given condition is true; unset otherwise.
func (c *CPU) SetC(condition bool) {
	if condition {
		c.c = 1
	} else {
		c.c = 0
	}
}

// SetZ will set the Z flag if the given condition is true; unset otherwise.
func (c *CPU) SetZ(condition bool) {
	if condition {
		c.z = 1
	} else {
		c.z = 0
	}
}

func (c *CPU) SetI(condition bool) {
	if condition {
		c.i = 1
	} else {
		c.i = 0
	}
}

func (c *CPU) SetD() {

}

func (c *CPU) SetB(condition bool) {
	if condition {
		c.b = 1
	} else {
		c.b = 0
	}
}

func (c *CPU) SetU(condition bool) {
	if condition {
		c.u = 1
	} else {
		c.u = 0
	}
}

// SetV will set the V flag if the given condition is true; unset otherwise.
func (c *CPU) SetV(condition bool) {
	if condition {
		c.v = 1
	} else {
		c.v = 0
	}
}

// SetN will set the N flag if the given value is negative; unset otherwise
func (c *CPU) SetN(value uint8) {
	if value&0x80 != 0 {
		c.n = 1
	} else {
		c.n = 0
	}
}
