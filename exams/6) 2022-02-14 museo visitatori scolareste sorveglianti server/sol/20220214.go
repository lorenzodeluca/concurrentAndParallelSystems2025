package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const N = 30   // capacità massima sala
const NC = 30  // massima capacità corridoio
const MaxS = 2 // massimo numero di sorveglianti dentro la sala
const DIMENSIONE_SCOLARESCA = 25

// info simulazione
const N_VISITATORI_SINGOLI = 10
const N_SCOLARESCHE = 2
const N_SORVEGLIANTI = 2

// altre variabili
const MAXBUFF = 100
const MAXPROC = N_VISITATORI_SINGOLI + N_SCOLARESCHE + N_SORVEGLIANTI //per sapere quanti proc in esecuzione oltre al server
const ID_VISITATORE_SINGOLO = 0
const ID_SCOLARESCA = 1
const ID_SORVEGLIANTI = 2
const N_UTENTI = 3
const ID_DIR_IN = 0
const ID_DIR_OUT = 1
const N_DIREZIONI = 2
const TEMPO_ATTESA_MAX = 1

// risorse/canali
var ingressoCorridoioIN [N_UTENTI]chan Richiesta // ingresso corridoio IN
var uscitaCorridoioIN chan Richiesta             // ingressoSala==uscitaCorridoioIn
var ingressoCorridoioOUT [N_UTENTI]chan Richiesta
var uscitaCorridoioOUT chan Richiesta

var closeSignalCmd = make(chan bool)
var endServerCmd = make(chan bool)
var done = make(chan bool)

// classi/strutture utili
type Richiesta struct {
	id   int //id utente
	tipo int
	ack  chan int
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

func idToString(tipo int) string {
	switch tipo {
	case ID_VISITATORE_SINGOLO:
		return "VISITATORE_SINGOLO"
	case ID_SCOLARESCA:
		return "SCOLARESCA"
	case ID_SORVEGLIANTI:
		return "SORVEGLIANTI"
	default:
		return "errore"
	}
}

func sleepRandTime(max int) {
	if max > 0 {
		time.Sleep(time.Duration(rand.Intn(max)) * time.Second)
	}
}

func cliente(id int, tipo int) {
	//per randomizzare i tempi di arrivo, opzionale
	sleepRandTime(TEMPO_ATTESA_MAX)

	fmt.Printf("[cliente %d/%s] è arrivato alla mostra\n", id, idToString(tipo))

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	//1. Imbocca il corridoio in direzione IN;
	fmt.Printf("[cliente %d/%s] 1. pronto a imboccare il corridoio in direzione IN;\n", id, idToString(tipo))
	ingressoCorridoioIN[tipo] <- ric
	<-ric.ack // waiting for server ack

	//2. Percorre il corridoio (direzione IN)
	fmt.Printf("[cliente %d/%s] 2. inizia a percorrere il corridoio (direzione IN);\n", id, idToString(tipo))
	sleepRandTime(TEMPO_ATTESA_MAX)

	//3. Entra nella sala della mostra abbandonando il corridoio;
	fmt.Printf("[cliente %d/%s] 3. pronto a entrare nella sala della mostra abbandonando il corridoio;\n", id, idToString(tipo))
	uscitaCorridoioIN <- ric
	<-ric.ack // waiting for server ack

	//4. Visita la mostra
	fmt.Printf("[cliente %d/%s] 4. Visita la mostra;\n", id, idToString(tipo))
	sleepRandTime(TEMPO_ATTESA_MAX)

	//5. Imbocca il corridoio in direzione OUT uscendo dalla sala;
	fmt.Printf("[cliente %d/%s] 5. è pronto ad uscire dalla mostra;\n", id, idToString(tipo))
	ingressoCorridoioOUT[tipo] <- ric
	<-ric.ack // waiting for server ack

	//6. Percorre il corridoio (direzione OUT)
	fmt.Printf("[cliente %d/%s] 6. inizia a percorrere il corridoio (direzione OUT)\n", id, idToString(tipo))
	sleepRandTime(TEMPO_ATTESA_MAX)

	//7. Esce dal corridoio in direzione OUT.
	fmt.Printf("[cliente %d/%s] è pronto ad uscire \n", id, idToString(tipo))
	uscitaCorridoioOUT <- ric
	<-ric.ack // waiting for server ack

	fmt.Printf("[cliente %d/%s] è uscito\n", id, idToString(tipo))
	done <- true
}

func addetto(id int) {
	tipo := ID_SORVEGLIANTI
	fmt.Printf("[sorvegliante %d] si è svegliato\n", id)

	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	richiestaUscita := 0
	for {
		//per randomizzare i tempi di arrivo, opzionale
		sleepRandTime(TEMPO_ATTESA_MAX)

		fmt.Printf("[sorvegliante %d] vuole entrare\n", id)

		var ric Richiesta
		ric.id = id
		ric.tipo = tipo
		ric.ack = make(chan int, MAXBUFF)

		//1. Imbocca il corridoio in direzione IN;
		fmt.Printf("[sorvegliante %d] 1. pronto a imboccare il corridoio in direzione IN;\n", id)
		ingressoCorridoioIN[tipo] <- ric
		richiestaUscita = <-ric.ack // waiting for server ack

		if richiestaUscita == -1 {
			break
		}

		//2. Percorre il corridoio (direzione IN)
		fmt.Printf("[sorvegliante %d] 2. inizia a percorrere il corridoio (direzione IN);\n", id)
		sleepRandTime(TEMPO_ATTESA_MAX)

		//3. Entra nella sala della mostra abbandonando il corridoio;
		fmt.Printf("[sorvegliante %d] 3. pronto a entrare nella sala della mostra abbandonando il corridoio;\n", id)
		uscitaCorridoioIN <- ric
		<-ric.ack // waiting for server ack

		//4. Visita la mostra
		fmt.Printf("[sorvegliante %d] 4. presiede la mostra;\n", id)
		sleepRandTime(TEMPO_ATTESA_MAX)

		//5. Imbocca il corridoio in direzione OUT uscendo dalla sala;
		fmt.Printf("[sorvegliante %d] 5. è pronto ad uscire dalla mostra;\n", id)
		ingressoCorridoioOUT[tipo] <- ric
		<-ric.ack // waiting for server ack

		//6. Percorre il corridoio (direzione OUT)
		fmt.Printf("[sorvegliante %d] 6. inizia a percorrere il corridoio (direzione OUT)\n", id)
		sleepRandTime(TEMPO_ATTESA_MAX)

		//7. Esce dal corridoio in direzione OUT.
		fmt.Printf("[sorvegliante %d] è pronto ad uscire \n", id)
		uscitaCorridoioOUT <- ric
		<-ric.ack // waiting for server ack

		fmt.Printf("[sorvegliante %d] è uscito\n", id)
	}
	fmt.Printf("[sorvegliante %d] ha finito\n", id)
	done <- true
}

// goroutine rappresentante il sistema di gestione del negozio
func server() {
	sorvegliantiInSala := 0
	clientiInSala := 0
	var personeInCorridoio [N_DIREZIONI]int //incluse scolaresche e sorveglianti
	var numSorvegliantiInCorridoioIngresso int
	var numScolarescheInCorridoio [N_DIREZIONI]int
	ciSonoUtentiInUscita := (len(ingressoCorridoioOUT[ID_SORVEGLIANTI]) + len(ingressoCorridoioOUT[ID_SCOLARESCA]) + len(ingressoCorridoioOUT[ID_VISITATORE_SINGOLO])) > 0
	//persone := make(map[int]Richiesta)
	closing := false

	for {
		/*
			da gestire
				var ingressoCorridoioIN [N_UTENTI]chan Richiesta // ingresso corridoio IN
				var uscitaCorridoioIN [N_UTENTI]chan Richiesta        // ingressoSala==uscitaCorridoioIn
				var ingressoCorridoioOUT [N_UTENTI]chan Richiesta
				var uscitaCorridoioOUT chan Richiesta
		*/
		personeInSalaECorridoioInTOT := sorvegliantiInSala + clientiInSala + personeInCorridoio[ID_DIR_IN]
		personeInCorridoioTOT := personeInCorridoio[ID_DIR_IN] + personeInCorridoio[ID_DIR_OUT]
		scolarescaIN := numScolarescheInCorridoio[ID_DIR_IN] > 0
		scolarescaOUT := numScolarescheInCorridoio[ID_DIR_OUT] > 0

		/* fmt.Printf("debugclosing:%t, \nsorvegliantiInSala:%d,\n clientiInSala:%d\n ,personeInCorridoio:%v\n, numScolarescheInCorridoio: %v\n", closing, sorvegliantiInSala, clientiInSala, personeInCorridoio, numScolarescheInCorridoio)
		for i := 0; i < N_UTENTI; i++ {
			fmt.Printf("debug1)%d\n", len(ingressoCorridoioIN[i]))
			fmt.Printf("debug2)%d\n", len(uscitaCorridoioIN[i]))
			fmt.Printf("debug3)%d\n", len(ingressoCorridoioOUT[i]))
			fmt.Printf("debug4)%d\n", len(uscitaCorridoioOUT[i]))
		} */

		select {

		//-----------------ingressoCorridoioIN:possono occupare il corridoio le persone solo se so già che ci staranno in sala

		// ingressoCorridoioIN[ID_SORVEGLIANTI], controllando capacità corridoio e sala e dipendenti, controllare se scolaresca in transito nel senso opposto, priorità utenti in uscita
		case ric := <-when(closing || (personeInSalaECorridoioInTOT+1 <= N && personeInCorridoioTOT+1 <= NC && sorvegliantiInSala+numSorvegliantiInCorridoioIngresso < MaxS && !scolarescaOUT && !ciSonoUtentiInUscita), ingressoCorridoioIN[ID_SORVEGLIANTI]):
			if closing {
				ric.ack <- -1
			} else {
				personeInCorridoio[ID_DIR_IN]++
				numSorvegliantiInCorridoioIngresso++
				ric.ack <- 1
			}

		// ingressoCorridoioIN[ID_VISITATORE_SINGOLO], sorv in sala, controllando capacità corridoio e sala, bisogna dare priorità a sorveglianti, controllare se scolaresca in transito nel senso opposto, priorità utenti in uscita
		case ric := <-when(sorvegliantiInSala > 0 && personeInSalaECorridoioInTOT+1 <= N && personeInCorridoioTOT+1 <= NC && !scolarescaOUT && len(ingressoCorridoioIN[ID_SORVEGLIANTI]) == 0 && !ciSonoUtentiInUscita, ingressoCorridoioIN[ID_VISITATORE_SINGOLO]):
			personeInCorridoio[ID_DIR_IN]++
			ric.ack <- 1

		// ingressoCorridoioIN[ID_SCOLARESCA], sorv in sala, controllando capacità corridoio e sala, bisogna dare priorità a sorveglianti e visitatori singoli, controllare se scolaresca in transito nel senso opposto, priorità utenti in uscita
		case ric := <-when(sorvegliantiInSala > 0 && personeInSalaECorridoioInTOT+DIMENSIONE_SCOLARESCA <= N && personeInCorridoioTOT+DIMENSIONE_SCOLARESCA <= NC && !scolarescaOUT && len(ingressoCorridoioIN[ID_SORVEGLIANTI]) == 0 && len(ingressoCorridoioIN[ID_VISITATORE_SINGOLO]) == 0 && !ciSonoUtentiInUscita, ingressoCorridoioIN[ID_SCOLARESCA]):
			personeInCorridoio[ID_DIR_IN] += DIMENSIONE_SCOLARESCA
			numScolarescheInCorridoio[ID_DIR_IN]++
			ric.ack <- 1

		//-----------------uscitaCorridoioIn
		case ric := <-uscitaCorridoioIN:
			switch ric.tipo {
			case ID_SORVEGLIANTI:
				personeInCorridoio[ID_DIR_IN]--
				numSorvegliantiInCorridoioIngresso--
				sorvegliantiInSala++
			case ID_VISITATORE_SINGOLO:
				personeInCorridoio[ID_DIR_IN]--
				clientiInSala++
			case ID_SCOLARESCA:
				personeInCorridoio[ID_DIR_IN] -= DIMENSIONE_SCOLARESCA
				numScolarescheInCorridoio[ID_DIR_IN]--
				clientiInSala += DIMENSIONE_SCOLARESCA
			}
			ric.ack <- 1

		//-----------------ingressoCorridoioOUT

		// ingressoCorridoioOUT[ID_SORVEGLIANTI], controllando capacità corridoio, controllare se scolaresca in transito nel senso opposto, priorità a visitatori singoli e scolaresche, solo se non si lascia la sala non sorvegliata
		case ric := <-when(personeInCorridoioTOT+1 <= NC && !scolarescaIN && len(ingressoCorridoioOUT[ID_VISITATORE_SINGOLO]) == 0 && len(ingressoCorridoioOUT[ID_SCOLARESCA]) == 0 && (sorvegliantiInSala > 1 || personeInSalaECorridoioInTOT-sorvegliantiInSala == 0), ingressoCorridoioOUT[ID_SORVEGLIANTI]):
			personeInCorridoio[ID_DIR_OUT]++
			sorvegliantiInSala--
			ric.ack <- 1

		// ingressoCorridoioOUT[ID_VISITATORE_SINGOLO], controllando capacità corridoio, controllare se scolaresca in transito nel senso opposto, dare priorità a scolaresche
		case ric := <-when(personeInCorridoioTOT+1 <= NC && !scolarescaIN && len(ingressoCorridoioOUT[ID_SCOLARESCA]) == 0, ingressoCorridoioOUT[ID_VISITATORE_SINGOLO]):
			personeInCorridoio[ID_DIR_OUT]++
			clientiInSala--
			ric.ack <- 1

		// ingressoCorridoioOUT[ID_SCOLARESCA], controllando capacità corridoio, controllare se scolaresca in transito nel senso opposto
		case ric := <-when(personeInCorridoioTOT+DIMENSIONE_SCOLARESCA <= NC && !scolarescaIN, ingressoCorridoioOUT[ID_SCOLARESCA]):
			personeInCorridoio[ID_DIR_OUT] += DIMENSIONE_SCOLARESCA
			clientiInSala -= 25
			numScolarescheInCorridoio[ID_DIR_OUT]++
			ric.ack <- 1

		//-----------------uscitaCorridoioOUT
		case ric := <-uscitaCorridoioOUT:
			switch ric.tipo {
			case ID_SCOLARESCA:
				personeInCorridoio[ID_DIR_OUT] -= DIMENSIONE_SCOLARESCA
				numScolarescheInCorridoio[ID_DIR_OUT]--
			default:
				personeInCorridoio[ID_DIR_OUT]--
			}
			ric.ack <- 1

		case <-closeSignalCmd:
			closing = true

		case <-endServerCmd:
			fmt.Println("THE END !!!!!!")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	/*
		var ingressoCorridoioIN [N_UTENTI]chan Richiesta // ingresso corridoio IN
		var uscitaCorridoioIN [N_UTENTI]chan Richiesta   // ingressoSala==uscitaCorridoioIn
		var ingressoCorridoioOUT [N_UTENTI]chan Richiesta
		var uscitaCorridoioOUT[N_UTENTI] chan Richiesta
	*/
	for i := 0; i < N_UTENTI; i++ {
		ingressoCorridoioIN[i] = make(chan Richiesta, MAXBUFF)
		ingressoCorridoioOUT[i] = make(chan Richiesta, MAXBUFF)
	}
	uscitaCorridoioIN = make(chan Richiesta, MAXBUFF)
	uscitaCorridoioOUT = make(chan Richiesta, MAXBUFF)

	go server()
	idProg := 0
	for i := 0; i < N_VISITATORI_SINGOLI; i++ {
		go cliente(idProg, ID_VISITATORE_SINGOLO)
		idProg++
	}
	for i := 0; i < N_SCOLARESCHE; i++ {
		go cliente(idProg, ID_SCOLARESCA)
		idProg++
	}
	for i := 0; i < N_SORVEGLIANTI; i++ {
		go addetto(idProg)
		idProg++
	}

	for i := 0; i < N_VISITATORI_SINGOLI+N_SCOLARESCHE; i++ {
		<-done
	}
	fmt.Printf("\n SERVER: INIZIO CHIUSURA \n")
	closeSignalCmd <- true
	for i := 0; i < N_SORVEGLIANTI; i++ {
		<-done
	}
	endServerCmd <- true
	<-done //waiting for server confirmation

	fmt.Printf("\n SERVER: HO FINITO ")
}
