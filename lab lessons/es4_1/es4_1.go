package main

import (
	"fmt"
	"math/rand"
	"time"
)

const MAXBUFF = 3
const MAXPROC = 3
const MAX = 10 // capacity
const N int = 0
const S int = 1
const BRIDGE_USER_TYPE_CAR int = 0
const BRIDGE_USER_TYPE_WALKER int = 1

var done = make(chan bool)
var endServerCmd = make(chan bool)
var entryCarGateN = make(chan int, MAXBUFF)    // north entry car queue
var entryCarGateS = make(chan int, MAXBUFF)    // south entry car queue
var entryWalkerGateN = make(chan int, MAXBUFF) // north entry Walker queue
var entryWalkerGateS = make(chan int, MAXBUFF) // south entry Walker queue
var exitCarGateN = make(chan int, MAXBUFF)     // north exit car queue
var exitCarGateS = make(chan int, MAXBUFF)     // south exit car queue
var exitWalkerGateN = make(chan int, MAXBUFF)  // north exit Walker queue
var exitWalkerGateS = make(chan int, MAXBUFF)  // south exit Walker queue
var ACK [MAXPROC]chan int                      //ack users
var r int

func when(b bool, c chan int) chan int {
	if !b {
		return nil
	}
	return c
}

func client(myid int) {
	var tt int = rand.Intn(2) + 1
	var userType int = rand.Intn(2) //rand between 0 inclusive - 2 esclusive
	var dir int = rand.Intn(2)      //rand between 0 inclusive - 2 esclusive

	fmt.Printf("[user/èPedone %d/%d]inizializzazione  %d direzione %d in secondi %d \n", myid, userType, myid, dir, tt)
	time.Sleep(time.Duration(tt) * time.Second)
	if dir == N { // asynchronous send
		if userType == BRIDGE_USER_TYPE_CAR {
			entryCarGateN <- myid
		} else {
			entryWalkerGateN <- myid
		}
		<-ACK[myid] // waiting server ack
		fmt.Printf("[user/èPedone %d/%d]  entrato sul ponte in direzione  NORD\n", myid, userType)
		tt = rand.Intn(5)
		time.Sleep(time.Duration(tt) * time.Second)
		if userType == BRIDGE_USER_TYPE_CAR {
			exitCarGateN <- myid
		} else {
			exitWalkerGateN <- myid
		}
		fmt.Printf("[user/èPedone %d/%d]  uscito dal ponte in direzione  NORD\n", myid, userType)
	} else {
		if userType == BRIDGE_USER_TYPE_CAR {
			entryCarGateS <- myid
		} else {
			entryWalkerGateS <- myid
		}
		<-ACK[myid] // waiting server ack
		fmt.Printf("[user/èPedone %d/%d]  entrato sul ponte in direzione  SOUTH\n", myid, userType)
		tt = rand.Intn(5)
		time.Sleep(time.Duration(tt) * time.Second)
		if userType == BRIDGE_USER_TYPE_CAR {
			exitCarGateS <- myid
		} else {
			exitWalkerGateS <- myid
		}
		fmt.Printf("[user/èPedone %d/%d]  uscito dal ponte in direzione  SOUTH\n", myid, userType)
	}
	done <- true
}

func server() {
	var cap int = 0
	var isCarOnTheBridgeN bool = false
	var isCarOnTheBridgeS bool = false
	var carsOnTheBridgeCounter = make(chan int, MAXBUFF)

	for {
		select {
		//to check for north walker entry: capacity, south walkers priority, cars on the opposite dir
		case x := <-when((cap < MAX) && (len(entryWalkerGateS) == 0) && !isCarOnTheBridgeS, entryWalkerGateN):
			cap++
			ACK[x] <- 1

		//to check for south walker entry: capacity, cars on the opposite dir
		case x := <-when((cap < MAX) && !isCarOnTheBridgeN, entryWalkerGateS):
			cap++
			ACK[x] <- 1

		//to check for south car entry: capacity, cars on the opposite dir, south walker priority
		case x := <-when((cap < MAX) && !isCarOnTheBridgeN && (len(entryWalkerGateS) == 0), entryCarGateS):
			cap += 10
			isCarOnTheBridgeS = true
			carsOnTheBridgeCounter <- x
			ACK[x] <- 1

		//to check for north car entry: capacity, cars on the opposite dir, south walker priority, south car priority, nord walker priority
		case x := <-when((cap < MAX) && !isCarOnTheBridgeS && (len(entryWalkerGateS) == 0) && (len(entryCarGateS) == 0) && (len(entryWalkerGateN) == 0), entryCarGateN):
			cap += 10
			isCarOnTheBridgeN = true
			carsOnTheBridgeCounter <- x
			ACK[x] <- 1

		case <-exitCarGateN:
			cap -= 10
			<-carsOnTheBridgeCounter
			if len(carsOnTheBridgeCounter) == 0 {
				isCarOnTheBridgeN = false
			}

		case <-exitCarGateS:
			cap -= 10
			<-carsOnTheBridgeCounter
			if len(carsOnTheBridgeCounter) == 0 {
				isCarOnTheBridgeS = false
			}

		case <-exitWalkerGateS:
			cap -= 1

		case <-exitWalkerGateN:
			cap -= 1

		case <-endServerCmd: // when all routines ended
			fmt.Println("FINE !!!!!!")
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

	for i := 0; i < MAXPROC; i++ {
		go client(i)
	}

	for i := 0; i < MAXPROC; i++ {
		<-done
	}
	endServerCmd <- true
	<-done
	fmt.Printf("\n HO FINITO ")
}
