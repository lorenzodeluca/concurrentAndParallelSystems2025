package main

import (
	"fmt"
	"math/rand"
	"time"
)

const NT = 3  // number of physiotherapists always inside the spa
const MAX = 3 // max number of users inside the spa
const USERS = 2
const LIFEGUARDS = 1
const MAXBUFF = 100
const MAXPROC = USERS + LIFEGUARDS

var closeCenterCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)
var entranceFunAreaLifeguard = make(chan int, MAXBUFF)
var entranceFunAreaUser = make(chan int, MAXBUFF)
var exitFunAreaLifeguard = make(chan int, MAXBUFF)
var exitFunAreaUser = make(chan int, MAXBUFF)
var entrancePhysiotherapyArea = make(chan int, MAXBUFF)
var exitPhysiotherapyArea = make(chan int, MAXBUFF)
var ACK [MAXPROC]chan int

func when(b bool, c chan int) chan int {
	if !b {
		return nil
	}
	return c
}

func user(id int) {
	fmt.Printf("[user %d] entered the spa center\n", id)
	for {
		//activity = 0 means the user wants to exit the center
		//activity = 1 means the user wants to enter the fun area
		//activity = 2 means the user wants to enter the fisio area
		var activity int = rand.Intn(3) // range 0-2 inclusive
		fmt.Printf("[user %d] choose the area %d\n", id, activity)
		if activity == 0 {
			break
		}

		//area entrance procedure
		if activity == 1 {
			entranceFunAreaUser <- id
		}
		if activity == 2 {
			entrancePhysiotherapyArea <- id
		}
		<-ACK[id] // waiting for server ack

		fmt.Printf("[user %d] entered the the area %d \n", id, activity)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		//area exit procedure
		if activity == 1 {
			exitFunAreaUser <- id
		}
		if activity == 2 {
			exitPhysiotherapyArea <- id
		}
		<-ACK[id] // waiting for server ack

		fmt.Printf("[user %d] left the the area %d \n", id, activity)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
	fmt.Printf("[user %d] left the spa center\n", id)
	done <- true
}

func lifeguard(id int) {
	fmt.Printf("[lifeguard %d] entered the spa center\n", id)
	for {
		fmt.Printf("[lifeguard %d] entering the fun area \n", id)

		//area entrance procedure
		entranceFunAreaLifeguard <- id

		if <-ACK[id] < 0 { // waiting for server ack
			//center closing
			break
		}

		fmt.Printf("[lifeguard %d] entered the the area \n", id)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		//area exit procedure
		exitFunAreaLifeguard <- id
		<-ACK[id] // waiting for server ack

		fmt.Printf("[lifeguard %d] left the the area \n", id)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
	fmt.Printf("[lifeguard %d] left the spa center\n", id)
	done <- true
}

// goroutine used to control the access to the center and its areas
func server() {
	capCounter := 0 //users inside center
	usersInsideFunAreaCounter := 0
	lifeguardsInsideFunAreaCounter := 0
	busyPhysiotherapistCounter := 0
	centerClosing := false

	for {
		select {
		// lifeguards always have priority
		case x := <-entranceFunAreaLifeguard:
			lifeguardsInsideFunAreaCounter++
			if centerClosing {
				ACK[x] <- -1
			} else {
				ACK[x] <- 1
			}

		//lifeguards can exit fun area only if there are no users or if there is another lifeguard
		case x := <-when(usersInsideFunAreaCounter == 0 || lifeguardsInsideFunAreaCounter > 1, exitFunAreaLifeguard):
			ACK[x] <- 1
			lifeguardsInsideFunAreaCounter--

		//users can enter fun area only if there is at least 1 lifeguard, they have priority over physiotherapy users. Lifeguards have priority over user entrance
		case x := <-when((capCounter < MAX) && len(entranceFunAreaLifeguard) == 0 && lifeguardsInsideFunAreaCounter > 0, entranceFunAreaUser):
			capCounter++
			ACK[x] <- 1

		//users entering Phys. area must give priority to the users entering the fun area, there must be an available Physiotherapist
		case x := <-when((capCounter < MAX) && len(entranceFunAreaUser) == 0 && NT-busyPhysiotherapistCounter > 0, entrancePhysiotherapyArea):
			capCounter++
			ACK[x] <- 1

		//users can always exit the fun area
		case x := <-exitFunAreaUser:
			ACK[x] <- 1
			capCounter--

		//users can always exit the physiotherapy area
		case x := <-exitPhysiotherapyArea:
			ACK[x] <- 1
			busyPhysiotherapistCounter--
			capCounter--

		case <-closeCenterCmd: // when all users ended
			centerClosing = true

		case <-endServerCmd: // when all routines(users+lifeguards) ended
			fmt.Println("THE END !!!!!!")
			done <- true
			return
		}
	}
}

func main() {

	//inizializzazione canali
	for i := 0; i < MAXPROC; i++ {
		ACK[i] = make(chan int, MAXBUFF)
	}

	go server()

	for i := 0; i < USERS; i++ {
		go user(i)
	}

	for i := 0; i < LIFEGUARDS; i++ {
		go lifeguard(i)
	}

	for i := 0; i < USERS; i++ {
		<-done
	}
	closeCenterCmd <- true
	for i := 0; i < LIFEGUARDS; i++ {
		<-done
	}
	endServerCmd <- true //telling server to shut down
	<-done               //waiting for server confirmation

	fmt.Printf("\n HO FINITO ")
}
