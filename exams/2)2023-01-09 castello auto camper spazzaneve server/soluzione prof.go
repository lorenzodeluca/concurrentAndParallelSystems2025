package main

import (
	"fmt"
	"math/rand"
	"time"
)

//tipi parcheggio:
const MAXI = 0
const STANDARD = 1

//tipologia veicoli:
const AUTOMOBILE = 0
const CAMPER = 1
const SPAZZANEVE = 2

const NS = 10 //posti "standard" parcheggio per automobile
const NM = 5  //posti "maxi" per automobile o camper

const N_TURISTI = 25 //camper e automobili
const MAXBUFF = 100

//direzione marcia
const SALITA = 0
const DISCESA = 1

//canali salita/discesa risp alla strada	(automobile = 0, camper = 1, spazzaneve = 2)
var inizioSalita [3]chan int
var fineSalita [3]chan int
var inizioDiscesa [3]chan Parcheggio
var fineDiscesa [3]chan int

//canali conferme operazioni
var ACK_tur [N_TURISTI]chan int
var ACK_spaz = make(chan int, MAXBUFF)

//canali terminazione
var done = make(chan bool)
var termina = make(chan bool)
var terminaSpazzaneve = make(chan bool)

type Parcheggio struct {
	index          int //id visitatore auto
	tipoParcheggio int //tipo parcheggio occupato (standard/maxi), interessante solo per automobili
}

func when(b bool, c chan int) chan int {
	if !b {
		return nil
	}
	return c
}

func whenParcheggio(b bool, c chan Parcheggio) chan Parcheggio {
	if !b {
		return nil
	}
	return c
}

func sleepRandTime(timeLimit int) {
	if timeLimit > 0 {
		time.Sleep(time.Duration(rand.Intn(timeLimit)+1) * time.Second)
	}
}

//Goroutine:

func turista(index int, tipo int) {

	var tipoParcheggio int

	sleepRandTime(4)

	inizioSalita[tipo] <- index
	tipoParcheggio = <-ACK_tur[index]

	sleepRandTime(3) // tempo di salita

	fineSalita[tipo] <- index
	<-ACK_tur[index]

	//visita castello
	sleepRandTime(4)

	p := Parcheggio{index, tipoParcheggio}

	inizioDiscesa[tipo] <- p
	<-ACK_tur[index]

	sleepRandTime(2)

	fineDiscesa[tipo] <- index
	<-ACK_tur[index]

	done <- true
	return
}

func spazzaneve() {
	var res int

	sleepRandTime(4)

	for {
		inizioDiscesa[SPAZZANEVE] <- Parcheggio{-1, -1}
		res = <-ACK_spaz
		if res == -1 {
			fmt.Printf("[spazzaneve] termino..\n")
			done <- true
			return
		}
		fmt.Printf("[spazzaneve] entrato in direzione DISCESA\n")
		sleepRandTime(2)

		fineDiscesa[SPAZZANEVE] <- 1
		res = <-ACK_spaz

		sleepRandTime(8)

		inizioSalita[SPAZZANEVE] <- 1
		res = <-ACK_spaz

		fmt.Printf("[spazzaneve] entrato in direzione SALITA\n")

		sleepRandTime(2)

		fineSalita[SPAZZANEVE] <- 1
		res = <-ACK_spaz

		fmt.Printf("[spazzaneve] entrato nel castello con successo!\n")

		sleepRandTime(8)
	}
}

func castello() {
	var index int
	var p Parcheggio
	var stop = false

	numCamperInStrada := [2]int{0, 0}
	numAutomobiliInStrada := [2]int{0, 0}
	var SpazzaneveInAzione = false

	var numStandard, numMaxi int = NS, NM //numero di posti liberi standard/maxi

	fmt.Printf("[castello] La strada è aperta ! \n")

	for {
		select {
		case index = <-when((numMaxi > 0) && (numCamperInStrada[DISCESA]+numAutomobiliInStrada[DISCESA] == 0) && (!SpazzaneveInAzione) &&
			(len(inizioDiscesa[CAMPER])+len(inizioDiscesa[AUTOMOBILE])+len(inizioDiscesa[SPAZZANEVE]) == 0), inizioSalita[CAMPER]):
			numMaxi--
			numCamperInStrada[SALITA]++
			fmt.Printf("[castello] entrato  CAMPER %d in direzione SALITA\n", index)
			ACK_tur[index] <- MAXI
		case index = <-when((numStandard+numMaxi > 0) && (numCamperInStrada[DISCESA] == 0) && (!SpazzaneveInAzione) && (len(inizioSalita[CAMPER]) == 0) &&
			(len(inizioDiscesa[CAMPER])+len(inizioDiscesa[AUTOMOBILE])+len(inizioDiscesa[SPAZZANEVE]) == 0), inizioSalita[AUTOMOBILE]):
			tipoParcheggio := -1
			if numStandard > 0 {
				numStandard--
				tipoParcheggio = STANDARD
			} else {
				numMaxi--
				tipoParcheggio = MAXI
			}
			numAutomobiliInStrada[SALITA]++
			fmt.Printf("[castello] entrata  AUTOMOBILE %d in  direzione SALITA\n", index)
			ACK_tur[index] <- tipoParcheggio
		case <-when((numCamperInStrada[DISCESA]+numAutomobiliInStrada[DISCESA]+numCamperInStrada[SALITA]+numAutomobiliInStrada[SALITA] == 0) &&
			(len(inizioSalita[CAMPER])+len(inizioSalita[AUTOMOBILE]) == 0) && (len(inizioDiscesa[CAMPER])+len(inizioDiscesa[AUTOMOBILE]) == 0), inizioSalita[SPAZZANEVE]):
			SpazzaneveInAzione = true
			fmt.Printf("[castello] entrato SPAZZANEVE in  direzione SALITA\n")
			ACK_spaz <- 1
		case index = <-fineSalita[CAMPER]:
			numCamperInStrada[SALITA]--
			fmt.Printf("[castello] entrato  CAMPER %d nel castello\n", index)
			ACK_tur[index] <- 1
		//automobile
		case index = <-fineSalita[AUTOMOBILE]:
			numAutomobiliInStrada[SALITA]--
			fmt.Printf("[castello] entrata  AUTOMOBILE %d nel castello\n", index)
			ACK_tur[index] <- 1
		//spazzaneve
		case <-fineSalita[SPAZZANEVE]:
			SpazzaneveInAzione = false
			fmt.Printf("[castello] entrato SPAZZANEVE nel castello\n")
			ACK_spaz <- 1
		case p = <-whenParcheggio((numCamperInStrada[SALITA]+numAutomobiliInStrada[SALITA] == 0) && (!SpazzaneveInAzione) && (len(inizioDiscesa[SPAZZANEVE]) == 0), inizioDiscesa[CAMPER]):
			numCamperInStrada[DISCESA]++
			numMaxi++
			fmt.Printf("[castello] entrato CAMPER %d in  direzione DISCESA\n", index)
			ACK_tur[p.index] <- 1
		case p = <-whenParcheggio((numCamperInStrada[SALITA] == 0) && (!SpazzaneveInAzione) && (len(inizioDiscesa[SPAZZANEVE])+len(inizioDiscesa[CAMPER]) == 0), inizioDiscesa[AUTOMOBILE]):
			numAutomobiliInStrada[DISCESA]++
			if p.tipoParcheggio == MAXI {
				numMaxi++
			} else {
				numStandard++
			}
			fmt.Printf("[castello] entrata AUTOMOBILE %d in direzione DISCESA\n", index)
			ACK_tur[p.index] <- 1
		case <-whenParcheggio((stop == false) && (numCamperInStrada[DISCESA]+numAutomobiliInStrada[DISCESA]+numCamperInStrada[SALITA]+numAutomobiliInStrada[SALITA] == 0), inizioDiscesa[SPAZZANEVE]):
			SpazzaneveInAzione = true
			fmt.Printf("[castello] entrato SPAZZANEVE in  direzione DISCESA\n")
			ACK_spaz <- 1
		case index = <-fineDiscesa[CAMPER]:
			numCamperInStrada[DISCESA]--
			fmt.Printf("[castello] uscito  CAMPER  %d\n", index)
			ACK_tur[index] <- 1
		case index = <-fineDiscesa[AUTOMOBILE]:
			numAutomobiliInStrada[DISCESA]--
			fmt.Printf("[castello] uscita AUTOMOBILE %d\n", index)
			ACK_tur[index] <- 1
		case <-fineDiscesa[SPAZZANEVE]:
			SpazzaneveInAzione = false
			fmt.Printf("[castello] uscito SPAZZANEVE\n")
			ACK_spaz <- 1
		//terminazione:
		case <-terminaSpazzaneve:
			stop = true
			fmt.Printf("[castello] turisti terminati, blocco spazzaneve..\n")
		case <-whenParcheggio((stop == true), inizioDiscesa[SPAZZANEVE]):
			ACK_spaz <- -1
		case <-termina: // quando lo spazzaneve ha terminato
			fmt.Printf("[castello] fine...\n")
			done <- true
			return
		}

	}
}

func main() {

	rand.Seed(time.Now().Unix())

	for i := 0; i < 3; i++ {
		inizioSalita[i] = make(chan int, MAXBUFF)
		fineSalita[i] = make(chan int, MAXBUFF)
		inizioDiscesa[i] = make(chan Parcheggio, MAXBUFF)
		fineDiscesa[i] = make(chan int, MAXBUFF)
	}

	for i := 0; i < N_TURISTI; i++ {
		ACK_tur[i] = make(chan int, MAXBUFF)
	}

	for i := 0; i < N_TURISTI; i++ {
		go turista(i, rand.Intn(2))
	}

	go spazzaneve()

	go castello()

	//attendo la fine dei turisti
	for i := 0; i < N_TURISTI; i++ {
		<-done
	}

	terminaSpazzaneve <- true
	<-done

	termina <- true
	<-done

	fmt.Printf("\n[MAIN] La strada è chiusa.\n")
}
