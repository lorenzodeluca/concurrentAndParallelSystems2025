package main

import (
	"fmt"
	"math/rand"
	"time"
)

const NREG = 20
const maxP = 300 //max vaccini VP nel deposito
const maxM = 100 //max vacciniVM nel deposito
const tipoVP = 0
const tipoVM = 1
const NL = 40     // numero dosi in un lotto
const Q = 50      //numero dosi richieste da ogni regione
const Ncicli = 10 //iterazioni per regione
const Rossa = 0
const Arancione = 1
const Gialla = 2
const TOTP = 10
const TOTM = 8

var tipoVaccino = [2]string{"Pfizer-BionTech", "Moderna"}
var tipoRegione = [20]string{"Valle D'aosta", "Piemonte", "Lombardia", "Veneto", "Friuli V.G.", "Trentino Alto Adige",
	"Liguria", "Toscana", "Emilia Romagna", "Marche", "Umbria", "Lazio", "Abruzzo", "Molise", "Campania", "Basilicata",
	"Puglia", "Calabria", "Sicilia", "Sardegna"}

var done = make(chan bool)
var termina = make(chan bool)
var terminaDeposito = make(chan bool)

//canali usati dalle regioni  per prelevare vaccini dal deposito
var prelievo [3]chan int //0->ross, 1->arancione, 2->gialla

//canali usati dalle aziende per prenotare/depositare lotti di vaccino nel deposito
var prenota [2]chan int
var consegna [2]chan int

//canali di ack
var ack_farm [2]chan int   // canale di ack per le case farmaceutiche
var ack_reg [NREG]chan int // canale di ack per le regioni

func when(b bool, c chan int) chan int {
	if !b {
		return nil
	}
	return c
}

func Farma(tipo int) {
	var tt int

	var goal int

	fmt.Printf("[Azienda %s]: partenza! \n", tipoVaccino[tipo])
	if tipo == tipoVP {
		goal = TOTP
	} else {
		goal = TOTM
	}

	for i := 0; i < goal; i++ { // per ogni ciclo..
		prenota[tipo] <- 1
		<-ack_farm[tipo]
		//fmt.Printf("[Azienda %s]: ciclo n. %d ho prenotato! \n", tipoVaccino[tipo], i)
		tt = (rand.Intn(4) + 1)
		time.Sleep(time.Duration(tt) * time.Second) //tempo di trasporto ..
		consegna[tipo] <- 1
		<-ack_farm[tipo]
		//fmt.Printf("[Azienda %s]: inserito Lotto n. %d di %d dosi\n", tipoVaccino[tipo], i+1, NL)
	}
	fmt.Printf("[Azienda %s]: terminato !\n", tipoVaccino[tipo])
	done <- true
	return
}

func regione(id int) {
	var zona, tt int

	fmt.Printf("[Regione %s]: partenza! \n", tipoRegione[id])

	for { // per ogni ciclo..
		zona = rand.Intn(3) // determino la zona
		//fmt.Printf("[Regione %s]: sono in zona %d! \n", tipoRegione[id], zona)
		prelievo[zona] <- id //richiedo Q dosi
		ris := <-ack_reg[id]
		if ris == -1 {
			fmt.Printf("[Regione %s]: sono costretta a terminare! \n", tipoRegione[id])
			done <- true
			return
		}
		tt = (rand.Intn(3) + 1)
		time.Sleep(time.Duration(tt) * time.Second) //tempo di passaggio al ciclo successivo..
		//fmt.Printf("[Regione %s]: prelevate %d dosi\n", tipoRegione[id], Q)
	}
	fmt.Printf("[Regione %s]: termino!\n", tipoRegione[id])
}

func deposito() {
	var inVP int = 0
	var inVM int = 0
	var prenotatiVP = 0
	var prenotatiVM = 0
	var reg, VPprel int
	var fine bool = false
	for {
		select {
		case <-when(inVP+prenotatiVP+NL <= maxP, prenota[tipoVP]):
			prenotatiVP += NL
			ack_farm[tipoVP] <- 1
			//fmt.Printf("[deposito]: prenotato lotto VP, ci sono %d VP prenotati e %d VP disponibili\n", prenotatiVP, inVP)
		case <-when((prenotatiVM+inVM+NL <= maxM), prenota[tipoVM]):
			prenotatiVM += NL
			ack_farm[tipoVM] <- 1
			//fmt.Printf("[deposito]: prenotato lotto VM, ci sono %d VM prenotati e %d VM disponibili\n", prenotatiVM, inVM)
		case <-consegna[tipoVP]:
			inVP += NL
			prenotatiVP -= NL
			ack_farm[tipoVP] <- 1
			fmt.Printf("[deposito]: consegnato lotto VP: ora ci sono %d VP  e %d VM disponibili\n", inVP, inVM)
		case <-consegna[tipoVM]:
			inVM += NL
			prenotatiVM -= NL
			ack_farm[tipoVM] <- 1
			fmt.Printf("[deposito]: consegnato lotto VM: ora ci sono %d VP  e %d VM disponibili\n", inVP, inVM)
		case reg = <-when(fine == false && (inVM+inVP >= Q), prelievo[Rossa]): //zona rossa
			if inVP >= Q {
				inVP -= Q
				VPprel = Q
			} else {
				VPprel = inVP
				inVP = 0
				inVM -= (Q - VPprel)
			}
			ack_reg[reg] <- 1
			fmt.Printf("[deposito]: regione %s in zona ROSSA ha prelevato %d vaccini VP e %d vaccini VM \n", tipoRegione[reg], VPprel, (Q - VPprel))
		case reg = <-when(fine == false && (inVM+inVP >= Q) && len(prelievo[Rossa]) == 0, prelievo[Arancione]): //zona arancione
			if inVP >= Q {
				inVP -= Q
				VPprel = Q
			} else {
				VPprel = inVP
				inVP = 0
				inVM -= (Q - VPprel)
			}
			ack_reg[reg] <- 1
			fmt.Printf("[deposito]: regione %s In zona ARANCIONE ha prelevato %d vaccini VP e %d vaccini VM \n", tipoRegione[reg], VPprel, (Q - VPprel))
		case reg = <-when(fine == false && (inVM+inVP >= Q) && len(prelievo[Rossa]) == 0 && len(prelievo[Arancione]) == 0, prelievo[Gialla]): //zona gialla
			if inVP >= Q {
				inVP -= Q
				VPprel = Q
			} else {
				VPprel = inVP
				inVP = 0
				inVM -= (Q - VPprel)
			}
			ack_reg[reg] <- 1
			fmt.Printf("[deposito]: regione %s In zona Gialla ha prelevato %d vaccini VP e %d vaccini VM \n", tipoRegione[reg], VPprel, (Q - VPprel))
		//terminazione
		case reg := <-when(fine == true, prelievo[Rossa]):
			ack_reg[reg] <- -1
		case reg := <-when(fine == true, prelievo[Arancione]):
			ack_reg[reg] <- -1
		case reg := <-when(fine == true, prelievo[Gialla]):
			ack_reg[reg] <- -1
		case <-termina:
			fine = true
		case <-terminaDeposito:
			fmt.Printf("[deposito]: termino (ci sono ancora %d dosi di VP e %d dosi di VM)\n", inVP, inVM)
			done <- true
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	fmt.Printf("[main] inizio..\n")
	// inizializzazioni:
	for i := 0; i < NREG; i++ {
		ack_reg[i] = make(chan int, 100)
	}
	for i := 0; i < 3; i++ {
		prelievo[i] = make(chan int, 100)
	}
	for i := 0; i < 2; i++ {
		prenota[i] = make(chan int, 100)
		consegna[i] = make(chan int, 100)
		ack_farm[i] = make(chan int, 100)
	}

	//creazione goroutine
	go deposito()
	go Farma(0) //Pfizer
	go Farma(1) //Moderna

	for i := 0; i < NREG; i++ {
		go regione(i)
	}
	// terminazione
	for i := 0; i < 2; i++ { //terminazione aziende farmaceutiche
		<-done
	}
	termina <- true

	for i := 0; i < NREG; i++ { //terminazione regioni
		<-done
	}

	terminaDeposito <- true
	<-done
	fmt.Printf("[main] APPLICAZIONE TERMINATA \n")
}
