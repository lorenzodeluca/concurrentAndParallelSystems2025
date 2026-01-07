package main

import (
	"fmt"
	"math/rand"
	"time"
)

const MAXBUFF = 100
const MAXPROC = 100

const Piccola = 0 //0.5 litri prezzo 0.10
const Grande = 1  //1.5 litri prezzo 0.20

const Csmall = 0.5 // capacità bottiglia piccola
const Cbig = 1.5   //capacità bottiglia grande

const N = 50.0 //capacità serbatoio (litri)

const C1 = 0 //cassetta bottiglie piccole
const C2 = 1 //cassetta bottiglie grandi

const M1 = 15 //max numero di monete da 10 cent
const M2 = 20 //max numero di monete da 20 cent

//canali dedicati ai clienti
var inizio_prelievo [2]chan richiesta
var fine_prelievo = make(chan richiesta, MAXBUFF)

//canali dedicati all' addetto
var inizio_rifornimento = make(chan int, MAXBUFF)
var fine_rifornimento = make(chan int, MAXBUFF)
var ACK_addetto = make(chan int, MAXBUFF)

//canali terminazione
var done = make(chan bool)
var termina = make(chan bool)
var terminaAddetto = make(chan bool)

type richiesta struct {
	index int
	tipo  int
	ack   chan int
}

func whenPrelievo(b bool, c chan richiesta) chan richiesta {
	if !b {
		return nil
	}
	return c
}

func when(b bool, c chan int) chan int {
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

//goroutine
func cittadino(index int) {
	tipo := rand.Intn(2) //0 per bottiglia piccola, 1 per bottiglia grande
	r := richiesta{index, tipo, make(chan int)}
	sleepRandTime(2) //simulazione pagamento
	if r.tipo == Piccola {
		fmt.Printf("[cittadino %d] richiesta di una bottiglia piccola\n", index)
	} else {
		fmt.Printf("[cittadino %d] richiesta di una bottiglia grande\n", index)
	}

	inizio_prelievo[tipo] <- r
	<-r.ack

	sleepRandTime(3) //erogazione
	fine_prelievo <- r
	<-r.ack
	fmt.Printf("[cittadino %d] ho riempito la mia bottiglia, termino !\n", index)
	done <- true
}

func addetto() {
	var res int
	sleepRandTime(4)

	for {
		inizio_rifornimento <- 1
		res = <-ACK_addetto
		if res == -1 {
			fmt.Printf("[addetto] termino..\n")
			done <- true
			return
		}

		fmt.Printf("[addetto] inizio rifornimento...\n")
		sleepRandTime(3) //tempo impiegato per rifornire

		fine_rifornimento <- 1
		<-ACK_addetto

		fmt.Printf("[addetto] Finito il rifornimento, casaH2O nuovamente disponibile...\n")
		sleepRandTime(5)
	}
}

func casaH2O() { //server

	var totH2O = N //capacità totale litri serbatoio
	var tot1 = 0   //totale monete presenti nella cassetta C1
	var tot2 = 0   //totale monete presenti nella cassetta C2
	var occupato = false
	var stop = false

	fmt.Printf("[casaH2O] casaH2O è in funzione!\n")

	for {
		select {
		//cittadino bottiglia piccola
		case x := <-whenPrelievo(!occupato && totH2O >= Csmall && tot1 < M1 &&
			(tot2 < M2 || (tot2 == M2 && len(inizio_rifornimento) == 0)), inizio_prelievo[Piccola]):
			occupato = true
			tot1++
			totH2O -= Csmall
			fmt.Printf("[casaH2O] il cittadino %d ha iniziato il prelievo di una bottiglia di tipo %d\n", x.index, x.tipo)
			x.ack <- 1
		//cittadino bottiglia grande
		case x := <-whenPrelievo(!occupato && totH2O >= Cbig && tot2 < M2 && len(inizio_prelievo[Piccola]) == 0 &&
			(tot1 < M1 || (tot1 == M1 && len(inizio_rifornimento) == 0)), inizio_prelievo[Grande]):
			occupato = true
			tot2++
			totH2O -= Cbig
			fmt.Printf("[casaH2O] il cittadino %d ha iniziato il prelievo di una bottiglia di tipo %d\n", x.index, x.tipo)
			x.ack <- 1
		//addetto rifornimento
		case <-when((stop == false) && !occupato &&
			((tot1 == M1 || tot2 == M2 || totH2O == 0) ||
				(len(inizio_prelievo[Piccola])+len(inizio_prelievo[Grande]) == 0)), inizio_rifornimento):
			occupato = true
			totH2O = N
			tot1 = 0
			tot2 = 0
			fmt.Printf("[casaH2O] l'addetto ha iniziato a riempire il serbatoio e svuotare le cassette\n")
			ACK_addetto <- 1
		//FINE PRELIEVI
		case x := <-fine_prelievo:
			occupato = false
			//fmt.Printf("[casaH2O] il cittadino %d ha finito il prelievo\n", x.index)
			x.ack <- 1
		case <-fine_rifornimento:
			occupato = false
			//fmt.Printf("[casaH2O] l'addetto ha terminato di effettuare la manutenzione\n")
			ACK_addetto <- 1
		//terminazione
		case <-terminaAddetto:
			stop = true
			fmt.Printf("[casaH2O] cittadini terminati, dico all'addetto di terminare\n")
		case <-when((stop == true), inizio_rifornimento):
			ACK_addetto <- -1
		case <-termina:
			fmt.Printf("[casaH2O] chiusura!\n")
			done <- true
			return
		}
	}

}

func main() {

	rand.Seed(time.Now().Unix())

	for i := 0; i < 2; i++ {
		inizio_prelievo[i] = make(chan richiesta, MAXBUFF)

	}

	for i := 0; i < MAXPROC; i++ {
		go cittadino(i)
	}

	go addetto()

	go casaH2O()
	fmt.Printf("\n[MAIN] casaH2O è aperta.\n")
	//attendo la fine dei turisti
	for i := 0; i < MAXPROC; i++ {
		<-done
	}

	terminaAddetto <- true
	<-done

	termina <- true
	<-done

	fmt.Printf("\n[MAIN] casaH2O è chiusa.\n")
}

/* riga 131: con la condizione (tot2<M2 || (tot2 == M2 && len(inizio_rifornimento) == 0)) indico il fatto che è possibile effettuare il prelievo della bottiglia
piccola nonostante la cassetta delle monete delle bottiglie grandi sia piena poichè non vi è alcun rifornimento in attesa

Discorso analogo per la riga 138*/
