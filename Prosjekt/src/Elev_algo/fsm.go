package Elev_algo

import(
    "time"
    "Driver"
    //"fmt"
    "encoding/json"
)

var (
    elevator Elevator
)

func setAllLights(e Elevator){
    for floor := 0; floor < Driver.NUMFLOORS; floor++{
        for btn := 0; btn < Driver.NUMBUTTONS; btn++{
            Driver.ElevSetButtonLight(btn, floor, e.Requests[floor][btn])
        }
    }
}

func fsm_onInitBetweenFloors(){
    Driver.ElevSetMotorDirection(int(D_Down))
    elevator.Dir = D_Down
    elevator.Behaviour = EB_Moving
    for Driver.ElevGetFloorSensorSignal() == -1{}
    Driver.ElevSetMotorDirection(int(D_Idle))
}

func fsm_onRequestButtonPress(btn_floor int, btn_type Button){
    switch(elevator.Behaviour){
    case EB_DoorOpen:
        if(elevator.Floor == btn_floor){
            timer_start(3000 * time.Millisecond)
        } else {
            elevator.Requests[btn_floor][btn_type] = true
        }
        break
    case EB_Moving:
        elevator.Requests[btn_floor][btn_type] = true
        break
    case EB_Idle:
        elevator.Requests[btn_floor][btn_type] = true
        elevator.Dir = requests_chooseDirection(elevator)
        if(elevator.Dir == D_Idle){
            Driver.ElevSetDoorLight(true)
            elevator = requests_clearAtCurrentFloor(elevator)
            timer_start(3000 * time.Millisecond)
            elevator.Behaviour = EB_DoorOpen
        } else {
            Driver.ElevSetMotorDirection(int(elevator.Dir))
            elevator.Behaviour = EB_Moving
        }
        break
    }
    
    setAllLights(elevator)
}


func fsm_onFloorArrival(newFloor int){
    elevator.Floor = newFloor
    Driver.ElevSetFloorIndicator(newFloor)
    switch(elevator.Behaviour){
    case EB_Moving:
        if(requests_shouldStop(elevator)){
            Driver.ElevSetMotorDirection(int(D_Idle))
            Driver.ElevSetDoorLight(true)
            elevator = requests_clearAtCurrentFloor(elevator)
            timer_start(3000 * time.Millisecond)
            setAllLights(elevator)
            elevator.Behaviour = EB_DoorOpen
        }
        break
    default:
        break
    }
}


func fsm_onDoorTimeout(){
    switch(elevator.Behaviour){
    case EB_DoorOpen:
        elevator.Dir = requests_chooseDirection(elevator)
        Driver.ElevSetMotorDirection(int(elevator.Dir))
        Driver.ElevSetDoorLight(false)
        if(elevator.Dir == D_Idle){
            elevator.Behaviour = EB_Idle
        } else {
            elevator.Behaviour = EB_Moving
        }
        break
    default:
        break
    }
}

func elevator_uninitialized() {
    elevator.Dir = D_Idle
    elevator.Behaviour = EB_Idle
    elevator.Floor = Driver.ElevGetFloorSensorSignal()
    for f := 0; f < Driver.NUMFLOORS; f++{
        for b := 0; b < Driver.NUMBUTTONS; b++{
            elevator.Requests[f][b] = false
        }
    }
}

func send_status(recieveCh chan []byte){
    for{
        time.Sleep(2000*time.Millisecond)
        var data []byte
        data = json.Marshal(elevator)
        recieveCh <- data
    }
}