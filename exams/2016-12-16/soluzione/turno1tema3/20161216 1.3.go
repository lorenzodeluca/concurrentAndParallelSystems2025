package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const NUM = 3

// info simulazione
const N_SPETTATORI = 20

// altre variabili
const MAXBUFF = 100
const MAXPROC = N_SPETTATORI //per sapere quanti proc in esecuzione oltre al server
const ID_TRIBUNA_LOCALE = 0
const ID_TRIBUNA_OSPITI = 1
const N_TRIBUNE = 2
const TEMPO_ATTESA_MAX = 5

// risorse/canali
var acquistoBiglietto chan Richiesta // int del ack sarà l'id del biglietto
var controlloVarco [N_TRIBUNE]chan Richiesta
var controlloVarcoCompletato chan Richiesta
var ingressoTribuna chan Richiesta
var mappaBiglietti = make(map[int]Biglietto) //key: id biglietto, value: info biglietto

var endBiglietteriaCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)

// classi/strutture utili
type Biglietto struct {
	idBiglietto  int //id
	idSpettatore int
	tribuna      int //ID_TRIBUNA_LOCALE/ID_TRIBUNA_OSPITI
}
type Richiesta struct {
	id  int //id utente
	ack chan int
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func idToString(tipo int) string {
	switch tipo {
	case ID_TRIBUNA_LOCALE:
		return "TRIBUNA_LOCALE"
	case ID_TRIBUNA_OSPITI:
		return "TRIBUNA_OSPITI ABITUALE"
	default:
		return "errore"
	}
}

func sleepRandTime(limitSecondi int) { //rand 0-limitSecondi inclusi
	if limitSecondi > 0 {
		time.Sleep(time.Duration(rand.Intn(limitSecondi)+1) * time.Second)
	}
}

func spettatore(id int) {
	var ric Richiesta
	ric.id = id
	ric.ack = make(chan int, 1)

	fmt.Printf("[spettatore %d] vuole entrare\n", id)

	//utente acquista biglietto
	acquistoBiglietto <- ric
	idBigliettoAcquistasto := <-ric.ack // waiting for server ack
	fmt.Printf("[spettatore %d] ha acquistato il biglietto\n", id)

	sleepRandTime(TEMPO_ATTESA_MAX)
	fmt.Printf("[spettatore %d] è arrivato al varco di sicurezza\n", id)

	//utente manda richiesta di controllo
	biglietto := mappaBiglietti[idBigliettoAcquistasto]
	controlloVarco[biglietto.tribuna] <- ric
	<-ric.ack // waiting for server ack

	fmt.Printf("[spettatore %d] controllo varco in corso...\n", id)
	sleepRandTime(TEMPO_ATTESA_MAX)
	fmt.Printf("[spettatore %d] controllo varco completato...\n", id)

	controlloVarcoCompletato <- ric
	<-ric.ack // waiting for server ack

	fmt.Printf("[spettatore %d] è stato autorizzato ad entrare nella tribuna %s\n", id, idToString(biglietto.tribuna))

	sleepRandTime(TEMPO_ATTESA_MAX)

	fmt.Printf("[spettatore %d] è arriva alla tribuna %s\n", id, idToString(biglietto.tribuna))
	ingressoTribuna <- ric
	<-ric.ack
	fmt.Printf("[spettatore %d] è entrato nella tribuna %s\n", id, idToString(biglietto.tribuna))

	done <- true
	fmt.Printf("[spettatore %d] ha finito \n", id)

}

func biglietteria() {
	idBigliettoProg := 0

	fmt.Printf("\n BIGLIETTERIA APERTA ")
	for {
		//condizioni

		/*
			da gestire
				var acquistoBiglietto chan Richiesta // int del ack sarà l'id del biglietto
		*/
		select {

		case ric := <-acquistoBiglietto:
			tribuna := rand.Intn(N_TRIBUNE)
			biglietto := Biglietto{
				idBiglietto:  idBigliettoProg,
				idSpettatore: ric.id,
				tribuna:      tribuna,
			}
			mappaBiglietti[biglietto.idBiglietto] = biglietto
			idBigliettoProg++
			ric.ack <- biglietto.idBiglietto

		case <-endBiglietteriaCmd:
			fmt.Printf("\n BIGLIETTERIA CHIUSA ")
			done <- true
			return
		}
	}
}

func stadio() {
	operatoriLiberi := NUM
	var contSpettEntranti [N_TRIBUNE]int
	for {
		prioritaLocali := contSpettEntranti[ID_TRIBUNA_LOCALE] >= contSpettEntranti[ID_TRIBUNA_OSPITI]
		prioritaOspiti := contSpettEntranti[ID_TRIBUNA_LOCALE] < contSpettEntranti[ID_TRIBUNA_OSPITI]

		select {

		/*
			da gestire
			var controlloVarco [N_TRIBUNE]chan Richiesta
			var controlloVarcoCompletato chan Richiesta
			var ingressoTribuna chan Richiesta
		*/
		// controlloVarco locali, priorità alla tribuna con maggiore utenti già entranti
		case ric := <-when((prioritaLocali || len(controlloVarco[ID_TRIBUNA_OSPITI]) == 0) && operatoriLiberi > 0, controlloVarco[ID_TRIBUNA_LOCALE]):
			operatoriLiberi--
			ric.ack <- 1

		// controlloVarco ospiti, priorità alla tribuna con maggiore utenti già entranti
		case ric := <-when((prioritaOspiti || len(controlloVarco[ID_TRIBUNA_LOCALE]) == 0) && operatoriLiberi > 0, controlloVarco[ID_TRIBUNA_OSPITI]):
			operatoriLiberi--
			ric.ack <- 1

		// ingresso/salita camper, non ci devono essere spazzaneve in transito,non ci devono essere camper in transito nel senso opposto, priorità code discesa, ci deve essere park maxi libero
		case ric := <-controlloVarcoCompletato:
			operatoriLiberi++
			ric.ack <- 1

		//ingressoTribuna
		case ric := <-ingressoTribuna:
			for _, infoBiglietto := range mappaBiglietti {
				if infoBiglietto.idSpettatore == ric.id {
					contSpettEntranti[infoBiglietto.tribuna]++
					break
				}
			}
			ric.ack <- 1

		case <-endServerCmd:
			fmt.Println("LO STADIO HA FINITO")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	/*
		var acquistoBiglietto chan Richiesta // int del ack sarà l'id del biglietto
		var controlloVarco [N_TRIBUNE]chan Richiesta
		var controlloVarcoCompletato chan Richiesta
		var ingressoTribuna chan Richiesta
	*/
	for i := 0; i < N_TRIBUNE; i++ {
		controlloVarco[i] = make(chan Richiesta, MAXBUFF)
	}
	acquistoBiglietto = make(chan Richiesta, MAXBUFF)
	controlloVarcoCompletato = make(chan Richiesta, MAXBUFF)
	ingressoTribuna = make(chan Richiesta, MAXBUFF)

	go stadio()
	go biglietteria()
	idProg := 0
	for i := 0; i < N_SPETTATORI; i++ {
		go spettatore(idProg)
		idProg++
	}

	for i := 0; i < N_SPETTATORI; i++ {
		<-done
	}
	endServerCmd <- true
	<-done //waiting for  confirmation
	endBiglietteriaCmd <- true
	<-done //waiting for  confirmation
	fmt.Printf("\n HO FINITO ")
}
