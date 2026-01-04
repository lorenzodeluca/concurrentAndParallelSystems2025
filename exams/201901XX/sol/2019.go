package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const MaxP = 5 // numero massimo di pneumatici (di qualunque tipo) che possono essere contemporaneamente stoccati nel deposito
const MaxC = 5 // numero massimo di cerchi (di qualunque tipo) che possono essere contemporaneamente stoccati nel deposito

// info simulazione
const TOT = 5              // auto da completare per la terminazione dell'applicazione
const TEMPO_ATTESA_MAX = 5 // tempo usato nei time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

// altre variabili
const MAXBUFF = 100
const MAXPROC = 4 + 2 //nastri + robot
const NUMERO_PRODOTTI = 4
const NUMERO_MODELLI = 2
const NUMERO_ID = NUMERO_PRODOTTI + NUMERO_MODELLI
const ID_MODELLO_A = 4 //modello auto
const ID_MODELLO_B = 5 //modello auto
const ID_TIPO_CA = 0   // identificativo prodotti in deposito: cerchione di tipo A
const ID_TIPO_CB = 1   // identificativo prodotti in deposito: cerchione di tipo B
const ID_TIPO_PA = 2   // identificativo prodotti in deposito: pneumatico di tipo A
const ID_TIPO_PB = 3   // identificativo prodotti in deposito: pneumatico di tipo B

// risorse/canali
var inserimentoInDeposito [NUMERO_PRODOTTI]chan Richiesta
var prelievoDaDeposito [NUMERO_PRODOTTI]chan Richiesta
var autorizzazioneInizioMacchina chan Richiesta
var notificaCompletamentoMacchina chan Richiesta

var closeSignalCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool) //per notificare la terminazione delle goroutine

// classi/strutture utili
type Richiesta struct {
	id   int //id utente
	tipo int // prodotto riguardante la richiesta
	ack  chan int
}

func getPneumaticoFromModello(modello int) int {
	switch modello {
	case ID_MODELLO_A:
		return ID_TIPO_PA
	case ID_MODELLO_B:
		return ID_TIPO_PB
	default:
		return -1
	}
}
func getCerchioFromModello(modello int) int {
	switch modello {
	case ID_MODELLO_A:
		return ID_TIPO_CA
	case ID_MODELLO_B:
		return ID_TIPO_CB
	default:
		return -1
	}
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func idToString(tipo int) string {
	switch tipo {
	case ID_TIPO_CA:
		return "CA"
	case ID_TIPO_CB:
		return "CB"
	case ID_TIPO_PA:
		return "PA"
	case ID_TIPO_PB:
		return "PB"
	case ID_MODELLO_A:
		return "A"
	case ID_MODELLO_B:
		return "B"
	default:
		return "errore: elemento non riconosciuto"
	}
}

func nastroTrasportatore(id int, tipo int) {
	fmt.Printf("[nastro %d/%s] avviato\n", id, idToString(tipo))

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	richiestaDiTerminazione := 0
	for {
		fmt.Printf("[nastro %d/%s] inizia il deposito\n", id, idToString(tipo))
		inserimentoInDeposito[tipo] <- ric
		richiestaDiTerminazione = <-ric.ack // waiting for server ack

		if richiestaDiTerminazione == -1 {
			break
		}
		fmt.Printf("[nastro %d/%s] deposito avvenuto con successo\n", id, idToString(tipo))
		time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

	}
	fmt.Printf("[nastro %d/%s] ha finito\n", id, idToString(tipo))
	done <- true
}

func robot(id int, modello int) {
	fmt.Printf("[robot %d/%s] avviato\n", id, idToString(modello))

	var ric Richiesta
	ric.id = id
	ric.tipo = modello
	ric.ack = make(chan int, MAXBUFF)

	richiestaUscita := 0
	for {
		ric.tipo = modello
		autorizzazioneInizioMacchina <- ric
		richiestaUscita = <-ric.ack
		if richiestaUscita == -1 {
			break
		}

		for i := 0; i < 4; i++ {
			ric.tipo = getCerchioFromModello(modello)
			prelievoDaDeposito[ric.tipo] <- ric
			<-ric.ack
			fmt.Printf("[robot %d/%s] montaggio cerchio %d\n", id, idToString(modello), i)

			//simulazione montaggio
			time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)

			ric.tipo = getPneumaticoFromModello(modello)
			prelievoDaDeposito[ric.tipo] <- ric
			<-ric.ack
			fmt.Printf("[robot %d/%s] montaggio pneumatico %d\n", id, idToString(modello), i)

			//simulazione montaggio
			time.Sleep(time.Duration(rand.Intn(TEMPO_ATTESA_MAX)) * time.Second)
		}
		ric.tipo = modello
		notificaCompletamentoMacchina <- ric
		<-ric.ack
	}
	fmt.Printf("[robot %d/%s] ha finito\n", id, idToString(modello))
	done <- true
}

func server() {
	var deposito [NUMERO_ID]int
	var conteggioAutoInLavorazione [NUMERO_ID]int

	//approx il numero di pezzi montati con il numero di pezzi richiesti per ridurre l'overhead del server nei meccanismi di sincronizzazione
	//contiene anche il numero di auto completate per modello
	var conteggioCompletamenti [NUMERO_ID]int

	closing := false
	autoCompletate := 0
	autoInLavorazione := 0

	for {
		/*
						da gestire
			var inserimentoInDeposito [DIMENSIONE_DEPOSITO]chan Richiesta
			var prelievoDaDeposito [DIMENSIONE_DEPOSITO]chan Richiesta
			var autorizzazioneInizioMacchina chan Richiesta
			var notificaCompletamentoMacchina chan Richiesta

		*/
		pneumaticiInDeposito := deposito[ID_TIPO_PA] + deposito[ID_TIPO_PB]
		cerchiInDeposito := deposito[ID_TIPO_CA] + deposito[ID_TIPO_CB]
		fmt.Printf("[SERVER/DEBUG]  deposito: %v\n", deposito)

		select {
		//per le immissioni devo garantire di avere in deposito almeno un pezzo per tipo altrimenti deadlock

		// immissione in deposito CA,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(closing || cerchiInDeposito < MaxC && (conteggioCompletamenti[ID_MODELLO_A] <= conteggioCompletamenti[ID_MODELLO_B] && (deposito[ID_TIPO_CB] != 0 || cerchiInDeposito < MaxC-1), inserimentoInDeposito[ID_TIPO_CA]):
			if closing {
				ric.ack <- -1
			} else {
				deposito[ric.tipo]++
				//consegna pezzo
				ric.ack <- 1
			}

		// immissione in deposito CB,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(closing || cerchiInDeposito < MaxC && (conteggioCompletamenti[ID_MODELLO_A] > conteggioCompletamenti[ID_MODELLO_B] || cerchiInDeposito == MaxC-1) && (deposito[ID_TIPO_CA] != 0 || cerchiInDeposito < MaxC-1), inserimentoInDeposito[ID_TIPO_CB]):
			if closing {
				ric.ack <- -1
			} else {
				deposito[ric.tipo]++
				//consegna pezzo
				ric.ack <- 1
			}

			// immissione in deposito PA,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(closing || pneumaticiInDeposito < MaxP && conteggioCompletamenti[ID_MODELLO_A] <= conteggioCompletamenti[ID_MODELLO_B] && (deposito[ID_TIPO_PB] != 0 || pneumaticiInDeposito < MaxP-1), inserimentoInDeposito[ID_TIPO_PA]):
			if closing {
				ric.ack <- -1
			} else {
				deposito[ric.tipo]++
				//consegna pezzo
				ric.ack <- 1
			}

			// immissione in deposito PB,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(closing || pneumaticiInDeposito < MaxP && (conteggioCompletamenti[ID_MODELLO_A] > conteggioCompletamenti[ID_MODELLO_B] || pneumaticiInDeposito == MaxP-1) && (deposito[ID_TIPO_PA] != 0 || pneumaticiInDeposito < MaxP-1), inserimentoInDeposito[ID_TIPO_PB]):
			if closing {
				ric.ack <- -1
			} else {
				deposito[ric.tipo]++
				//consegna pezzo
				ric.ack <- 1
			}

			//------------------------prelievi
			// prelievo in deposito CA,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(deposito[ID_TIPO_CA] > 0 && conteggioCompletamenti[ID_MODELLO_A] <= conteggioCompletamenti[ID_MODELLO_B], prelievoDaDeposito[ID_TIPO_CA]):
			deposito[ric.tipo]--
			conteggioCompletamenti[ric.tipo]++
			//consegna pezzo
			ric.ack <- 1

		// prelievo in deposito CB,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(deposito[ID_TIPO_CB] > 0 && conteggioCompletamenti[ID_MODELLO_A] >= conteggioCompletamenti[ID_MODELLO_B], prelievoDaDeposito[ID_TIPO_CB]):
			deposito[ric.tipo]--
			conteggioCompletamenti[ric.tipo]++
			//consegna pezzo
			ric.ack <- 1

			// prelievo in deposito PA,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(deposito[ID_TIPO_PA] > 0 && conteggioCompletamenti[ID_MODELLO_A] <= conteggioCompletamenti[ID_MODELLO_B], prelievoDaDeposito[ID_TIPO_PA]):
			deposito[ric.tipo]--
			conteggioCompletamenti[ric.tipo]++
			//consegna pezzo
			ric.ack <- 1

			// prelievo in deposito PB,precedenza al  modello di auto con il minor numero di montaggi ruote completati
		case ric := <-when(deposito[ID_TIPO_PB] > 0 && conteggioCompletamenti[ID_MODELLO_A] >= conteggioCompletamenti[ID_MODELLO_B], prelievoDaDeposito[ID_TIPO_PB]):
			deposito[ric.tipo]--
			conteggioCompletamenti[ric.tipo]++
			//consegna pezzo
			ric.ack <- 1

		case ric := <-autorizzazioneInizioMacchina:
			if autoInLavorazione+autoCompletate < TOT {
				// autorizzazione inizio auto
				conteggioAutoInLavorazione[ric.tipo]++
				autoInLavorazione = conteggioAutoInLavorazione[ID_MODELLO_A] + conteggioAutoInLavorazione[ID_MODELLO_B]
				ric.ack <- 1
			} else {
				// spegnimento robot
				ric.ack <- -1
			}

		case ric := <-notificaCompletamentoMacchina:
			fmt.Printf("[SERVER/DEBUG] %d AUTO COMPLETATA\n", autoCompletate)
			conteggioAutoInLavorazione[ric.tipo]--
			conteggioCompletamenti[ric.tipo]++
			autoInLavorazione = conteggioAutoInLavorazione[ID_MODELLO_A] + conteggioAutoInLavorazione[ID_MODELLO_B]
			autoCompletate = conteggioCompletamenti[ID_MODELLO_A] + conteggioCompletamenti[ID_MODELLO_B]
			//se non ci sono altre auto in lavorazione chiudo il centro
			if autoInLavorazione == 0 && autoCompletate >= TOT {
				closing = true
			}
			ric.ack <- 1

		case <-endServerCmd: // when all routines ended
			fmt.Println("THE END !!!!!!")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	/*
		var inserimentoInDeposito [NUMERO_PRODOTTI]chan Richiesta
		var prelievoDaDeposito [NUMERO_PRODOTTI]chan Richiesta
		var autorizzazioneInizioMacchina chan Richiesta
		var notificaCompletamentoMacchina chan Richiesta
	*/
	for i := 0; i < NUMERO_PRODOTTI; i++ {
		inserimentoInDeposito[i] = make(chan Richiesta, MAXBUFF)
		prelievoDaDeposito[i] = make(chan Richiesta, MAXBUFF)
	}
	autorizzazioneInizioMacchina = make(chan Richiesta, MAXBUFF)
	notificaCompletamentoMacchina = make(chan Richiesta, MAXBUFF)

	go server()
	idProg := 0
	for i := 0; i < NUMERO_PRODOTTI; i++ {
		go nastroTrasportatore(idProg, i)
		idProg++
	}
	go robot(0, ID_MODELLO_A)
	go robot(1, ID_MODELLO_B)

	for i := 0; i < MAXPROC; i++ {
		<-done //waiting for server confirmation
	}

	fmt.Printf("\n HO FINITO ")
}
