package Elev_control

import (
	"Driver"
	"fmt"
	"time"
)

type Elevator struct {
	Floor     int
	Dir       Direction
	Requests  [Driver.NUMFLOORS][Driver.NUMBUTTONS]bool
	Behaviour ElevatorBehaviour
	Elev_ID   int64
	//doorOpenDuration_s float
}

var (
	elevator      Elevator
	allExtBtns    [Driver.NUMFLOORS][Driver.NUMBUTTONS - 1]bool
	lastFloorTime int64
)

//var requests_timeStamp [4][3]int64 //Setter timeStamp når en ordre aktiveres.

func Run_Elevator(localStatusCh chan Elevator, sendBtnCallCh chan [2]int, receiveAllBtnCallsCh chan [Driver.NUMFLOORS][Driver.NUMBUTTONS - 1]bool, setLights_setExtBtnsCh chan [4][2]bool, errorCh chan int) {
	//var (
	//lastFloorTime int64
	//)
	//Init elev_state
	if Driver.ElevGetFloorSensorSignal() == -1 {
		fsm_onInitBetweenFloors()
	}
	fsm_elevatorUninitialized()
	fmt.Println("Elev ID: ", elevator.Elev_ID)

	running := true
	var prev_button [Driver.NUMFLOORS][Driver.NUMBUTTONS]int
	var prev_floor int
	prev_floor = Driver.ElevGetFloorSensorSignal()
	go Update_ExtBtnCallsInElevControl(setLights_setExtBtnsCh) //SETTE DENNE INN I RUNNING MEN UTEN FORLØKKE INNI SEG SELV?
	//go checkElevMoving(errorCh) kan sette på senere
	//go updateAllExtLights(receiveAllBtnCallsCh)
	update_status_count := 0
	update_lights_count := 0
	localStatusCh <- elevator
	for running {
		// Request button
		for f := 0; f < Driver.NUMFLOORS; f++ {
			for b := 0; b < Driver.NUMBUTTONS; b++ {
				v := Driver.ElevGetButtonSignal(b, f)
				if v&int(v) != prev_button[f][b] {
					if fsm_onRequestButtonPress(f, Button(b), sendBtnCallCh) { //Hvis true er det innvendig knappetrykk
						temp_Behaviour := elevator.Behaviour
						fsm_onNewActiveRequest(f, Button(b))
						if temp_Behaviour != elevator.Behaviour {
							localStatusCh <- elevator
						}
					} else {
						fsm_SendNewOrderToMaster(f, Button(b), sendBtnCallCh)
					}
				}
				prev_button[f][b] = v
			}
		}
		// Floor sensor
		f := Driver.ElevGetFloorSensorSignal()
		if f != -1 {
			if f != prev_floor {
				fsm_onFloorArrival(f)
				localStatusCh <- elevator
				fmt.Println("\nAnkommer nytt floor: elevator.Floor: ", elevator.Floor, "\n")
			}
		}
		prev_floor = f
		// Timer
		if timer_timedOut() {
			fsm_onDoorTimeout()
			timer_stop()
			localStatusCh <- elevator
		}
		if update_status_count == 100 { //HVA ER GREIA MED DENNE? GJØRE DEN KORTERE?
			localStatusCh <- elevator
			//setLights_setExtBtnsCh <- allExtBtns
			update_status_count = 0
		}
		if update_lights_count == 10 {
			setLights_setExtBtnsCh <- allExtBtns
			update_lights_count = 0
		}
		time.Sleep(25 * time.Millisecond) //Hardkoding
		update_status_count += 1
		update_lights_count += 1
	}
}

/*func checkElevMoving(errorCh chan int) {
	var errorTime int64
	var timeNow int64
	errorTime = 8
	err := 1
	for {
		time.Sleep(1 * time.Second)
		if elevator.Behaviour == EB_Moving {
			timeNow = time.Now().Unix()
			if timeNow-lastFloorTime > errorTime {
				errorCh <- err
			}
		}
	}
}*/

func Update_ExtBtnCallsInElevControl(setLights_setExtBtnsCh chan [4][2]bool) {
	var temp_allExtBtns [4][2]bool
	for {
		temp_allExtBtns = <-setLights_setExtBtnsCh
		allExtBtns = temp_allExtBtns
		setAllLights()
	}
}

func PrintElev(elev Elevator) {
	fmt.Println("")
	fmt.Println("Floor: ", elev.Floor)
	fmt.Println("Direction: ", elev.Dir)
	for f := 0; f < 4; f++ {
		fmt.Printf("%+v", elev.Requests[f])
		fmt.Println("")
	}
}
