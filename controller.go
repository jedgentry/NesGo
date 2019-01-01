package main

const (
	//ButtonA on the gamepad.
	ButtonA = iota
	//ButtonB on the gamepad.
	ButtonB
	//ButtonSelect on the gamepad.
	ButtonSelect
	//ButtonStart on the gamepad.
	ButtonStart
	//ButtonUp on the gamepad.
	ButtonUp
	//ButtonDown on the gamepad.
	ButtonDown
	//ButtonLeft on the gamepad.
	ButtonLeft
	//ButtonRight on the gamepad.
	ButtonRight
)

//Controller represents a users controller.
type Controller struct {
	buttons [8]bool
	index   int
	strobe  byte
}

func (system *System) resetControllers() {
	for i := 0; i < len(system.controller); i++ {
		system.controller[i].resetController()
	}
}

func (controller *Controller) resetController() {

}

func (controller *Controller) Read() byte {
	data := byte(0)
	if controller.index < 8 && controller.buttons[controller.index] {
		data |= 1
	}
	if controller.strobe&1 == 1 {
		controller.index = 0
	} else {
		controller.index++
	}
	return data
	// XXX simulate open bus
}

func (controller *Controller) Write(data byte) {
	controller.strobe = data
	if controller.strobe&1 == 1 {
		controller.index = 0
	}
}
