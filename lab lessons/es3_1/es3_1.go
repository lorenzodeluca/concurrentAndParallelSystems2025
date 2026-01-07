package main

import (
	"fmt"
	"math/rand"
)

const MAXPROC = 5
const MAXBT = 2
const MAXEB = 1

type Bike struct {
	id       int //id of the proc that can use this bike, //<0 means bike id TODO, >0  means proc id
	bikeType int // 𝑩𝑻=0, 𝑬𝑩=1, 𝑭𝑳𝑬X=2   -----> FLEX ONLY FOR REQUEST
}

/*
valori nel canale
𝑩𝑻=0, 𝑬𝑩=1, 𝑭𝑳𝑬X=2
*/
var richiesta = make(chan Bike)

/*
valori nel canale
𝑩𝑻=0, 𝑬𝑩=1
*/
var rilascio = make(chan Bike)

/*
valori nel canale
𝑩𝑻=0, 𝑬𝑩=1
*/
var risorsa [MAXPROC]chan Bike
var done = make(chan int)
var termina = make(chan int)

func server() {
	fmt.Println("S: server avviato")
	//init available bikes
	var r, b Bike
	availableBT := make(chan Bike, MAXBT)
	availableEB := make(chan Bike, MAXEB)
	waitingBT := make(chan Bike, MAXPROC)   //if a client requested a bike and it wasnt available, the pending request is stored here
	waitingEB := make(chan Bike, MAXPROC)   //if a client requested a e-bike and it wasnt available, the pending request is stored here
	waitingFLEX := make(chan Bike, MAXPROC) //if a client requested a flex and it wasnt available, the pending request is stored here

	for i := 0; i < MAXBT; i++ {
		availableBT <- Bike{-1, 0}
	}
	for i := 0; i < MAXEB; i++ {
		availableEB <- Bike{-1, 1}
	}

	for {
		select {
		case r = <-richiesta:
			fmt.Println("S: richiesta", r.id, " per ", r.bikeType)
			if r.bikeType == 0 {
				if len(availableBT) > 0 {
					risorsa[r.id] <- <-availableBT
				} else {
					waitingBT <- r
					fmt.Println("S: richiesta", r.id, " aggiunta alla coda delle attese BT")

				}

			} else if r.bikeType == 1 {
				if len(availableEB) > 0 {
					risorsa[r.id] <- <-availableEB
				} else {
					waitingEB <- r
					fmt.Println("S: richiesta", r.id, " aggiunta alla coda delle attese EB")
				}
			} else {
				if len(availableEB) > 0 {
					risorsa[r.id] <- <-availableEB
				} else if len(availableBT) > 0 {
					risorsa[r.id] <- <-availableBT
				} else {
					waitingFLEX <- r
					fmt.Println("S: richiesta", r.id, " aggiunta alla coda delle attese FLEX")
				}
			}
		case b = <-rilascio:
			fmt.Println("S: restituita bici", r.id)
			if b.bikeType == 0 {
				if len(waitingBT) > 0 {
					clientReq := <-waitingBT
					risorsa[clientReq.id] <- b
				} else if len(waitingFLEX) > 0 {
					clientReq := <-waitingFLEX
					risorsa[clientReq.id] <- b
				} else {
					availableBT <- b
				}
			} else {
				if len(availableEB) > 0 {
					clientReq := <-waitingEB
					risorsa[clientReq.id] <- b
				} else if len(waitingFLEX) > 0 {
					clientReq := <-waitingFLEX
					risorsa[clientReq.id] <- b
				} else {
					availableEB <- b
				}
			}

		case <-termina: // end
			done <- 1 //server end request
			return
		}
	}
}

func client(i int) {
	fmt.Println("C-", i, " client avviato")
	richiesta <- Bike{i, rand.Intn(3)} //rand between 0 inclusive - 3 esclusive -> 𝑩𝑻=0, 𝑬𝑩=1, 𝑭𝑳𝑬X=2
	r := <-risorsa[i]                  //waiting that the requested bike is available
	fmt.Println("C-", i, " ottenuta bici")
	rilascio <- r
	done <- i
}

func main() {

	//init
	for i := 0; i < MAXPROC; i++ {
		risorsa[i] = make(chan Bike)
	}

	go server()
	for i := 0; i < MAXPROC; i++ {
		go client(i)
	}

	for i := 0; i < MAXPROC; i++ {
		fmt.Println("MAIN: HA TERMINATO", <-done)
	}
	termina <- 1
	<-done
	fmt.Println("finito!")
}
