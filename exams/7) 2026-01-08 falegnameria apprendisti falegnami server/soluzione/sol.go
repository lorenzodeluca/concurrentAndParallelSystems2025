// Lorenzo De Luca 0001240461
package main

import (
	"fmt"
	"math/rand"
	"time"
)

// limiti
const N_S = 50 // numero seghe circolari
const N_F = 10 // fresatrici elettriche
const ES = 3   // potenza in Kw della sega circolare
const EF = 3   // potenza in Kw della per ogni fresatrice
const MAX = 12 // max potenza in Kw del laboratorio

// info simulazione
const N_OPERAI = 5                // numero di operai che saranno fatti partire per la simulazione con tipologia(falegname/apprendista) casuale
const N_LAVORAZIONI_PER_TURNO = 2 // lavorazioni che l'operaio deve fare prima di lasciare la fabbrica

// altre variabili
const MAXBUFF = 100      //dimensione massima buffer canali
const MAXPROC = N_OPERAI //per sapere quanti proc in esecuzione oltre al server

const ID_FALEGNAME = 0
const ID_APPRENDISTA = 1
const N_TIPOLOGIE_OPERAI = 2

const ID_SEGA_CIRCOLARE = 0
const ID_FRESATRICE_ELETTRICA = 1
const N_TIPOLOGIE_STRUMENTI = 2

const TEMPO_ATTESA_MAX = 0 // il tempo massimo che verranno messe in pausa le goroutine quando bisogna simulare un attesa

// risorse/canali
var acquisizioneStrumento [N_TIPOLOGIE_OPERAI][N_TIPOLOGIE_STRUMENTI]chan Richiesta //canali asincroni
var rilascioStrumento chan Richiesta                                                //canale asincrono

var endServerCmd = make(chan bool) // canale sincrono
var done = make(chan bool)         // canale sincrono

// classi/strutture utili
type Richiesta struct {
	id   int      //id utente
	tipo int      //tipo strumento
	ack  chan int //canale asincrono per la sincronizzazione, asincrono per non far attendere il server in caso l'operaio non sia pronto a ricevere il messaggio
}

func when(b bool, c chan Richiesta) chan Richiesta {
	if !b {
		return nil
	}
	return c
}

// data una tipologia di operaio(int) restituisce il nome della tipologia(string)
func idOperaioToString(tipo int) string {
	switch tipo {
	case ID_FALEGNAME:
		return "FALEGNAME"
	case ID_APPRENDISTA:
		return "APPRENDISTA"
	default:
		return "errore idOperaioToString"
	}
}

// data una tipologia di strumento(int) restituisce il nome della tipologia(string)
func idStrumentoToString(tipo int) string {
	switch tipo {
	case ID_SEGA_CIRCOLARE:
		return "SEGA CIRCOLARE"
	case ID_FRESATRICE_ELETTRICA:
		return "FRESATRICE ELETTRICA"
	default:
		return "errore idStrumentoToString"
	}
}

// mette in pausa la goroutine chiamante per max(int) secondi
func sleepRandTime(max int) {
	if max > 0 {
		time.Sleep(time.Duration(rand.Intn(max)) * time.Second)
	}
}

// goroutine rappresentante gli operai, falegnami se tipo vale ID_FALEGNAME e apprendisti se tipo vale ID_APPRENDISTI
func operaio(id int, tipo int) {
	var ric Richiesta
	ric.id = id
	ric.tipo = tipo
	ric.ack = make(chan int, MAXBUFF)

	//sleepRandTime(TEMPO_ATTESA_MAX) //per randomizzare i tempi di arrivo
	fmt.Printf("[operaio %d/%s] è entrato in fabbrica\n", id, idOperaioToString(tipo))

	for lavorazioniEffettuate := 0; lavorazioniEffettuate < N_LAVORAZIONI_PER_TURNO; lavorazioniEffettuate++ {
		// 1. Acquisizione dello strumento

		//selezione casuale dello strumento da usare
		idStrumento := rand.Intn(N_TIPOLOGIE_STRUMENTI) // genera tra 0 incluso e N_TIPOLOGIE_STRUMENTI escluso

		fmt.Printf("[operaio %d/%s] richiede di usare la %s\n", id, idOperaioToString(tipo), idStrumentoToString(idStrumento))

		//invio della richiesta di acquisizione al server
		ric.tipo = idStrumento
		acquisizioneStrumento[tipo][idStrumento] <- ric
		<-ric.ack // attesa ack

		// 2. < uso dello strumento>
		fmt.Printf("[operaio %d/%s] lavora usando la %s\n", id, idOperaioToString(tipo), idStrumentoToString(idStrumento))
		sleepRandTime(TEMPO_ATTESA_MAX) //simulazione dell'utilizzo dello strumento

		//3. Rilascio dello strumento
		fmt.Printf("[operaio %d/%s] ha finito di usare la %s\n", id, idOperaioToString(tipo), idStrumentoToString(idStrumento))

		//richiesta di riconsegna dello strumento
		rilascioStrumento <- ric
		<-ric.ack // attesa della conferma di ricezione da parte del server

		fmt.Printf("[operaio %d/%s] ha rilasciato la %s\n", id, idOperaioToString(tipo), idStrumentoToString(idStrumento))
		sleepRandTime(TEMPO_ATTESA_MAX) //simulazione del lavoro dell'operaio finchè non si prepara a richiedere un nuovo strumento
	}
	fmt.Printf("[operaio %d/%s] è uscito dalla fabbrica\n", id, idOperaioToString(tipo))
	done <- true
}

// goroutine rappresentante il sistema di gestione della falegnameria
func server() {
	var strumentiLiberi [N_TIPOLOGIE_STRUMENTI]int
	strumentiLiberi[ID_SEGA_CIRCOLARE] = N_S
	strumentiLiberi[ID_FRESATRICE_ELETTRICA] = N_F

	var kwCheConsuma [N_TIPOLOGIE_STRUMENTI]int
	kwCheConsuma[ID_SEGA_CIRCOLARE] = ES
	kwCheConsuma[ID_FRESATRICE_ELETTRICA] = EF

	var kwUtilizzati = 0

	for {
		//variabili ausiliare/condizioni

		//prioritaSeghe(boolean): se le seghe sono prioritarie
		prioritaSeghe := strumentiLiberi[ID_SEGA_CIRCOLARE] >= strumentiLiberi[ID_FRESATRICE_ELETTRICA]

		//prioritaFresatrici(boolean): se le fresatrici sono prioritarie
		prioritaFresatrici := strumentiLiberi[ID_FRESATRICE_ELETTRICA] >= strumentiLiberi[ID_SEGA_CIRCOLARE]

		//codaSeghe(int):quanti operai sono in coda per ottenere una sega
		codaSeghe := len(acquisizioneStrumento[ID_FALEGNAME][ID_SEGA_CIRCOLARE]) + len(acquisizioneStrumento[ID_APPRENDISTA][ID_SEGA_CIRCOLARE])

		//codaSegheFalegnami(int):quanti falegnami sono in coda per ottenere una sega
		codaSegheFalegnami := len(acquisizioneStrumento[ID_FALEGNAME][ID_SEGA_CIRCOLARE])

		//codaFresatriciFalegnami(int):quanti falegnami sono in coda per ottenere una fresatrice
		codaFresatriciFalegnami := len(acquisizioneStrumento[ID_FALEGNAME][ID_FRESATRICE_ELETTRICA])

		//codaFresatrici(int):quanti operai sono in coda per ottenere una fresatrice
		codaFresatrici := len(acquisizioneStrumento[ID_FALEGNAME][ID_FRESATRICE_ELETTRICA]) + len(acquisizioneStrumento[ID_APPRENDISTA][ID_FRESATRICE_ELETTRICA])

		// output di debug (opzionale)
		//fmt.Printf("\n\n -----DEBUG STATO SERVER----\nkwUtilizzati:%d\n,kwCheConsuma:%v\nprioritaFresatrici:%t\nprioritaSeghe:%t\nStrumenti Liberi: %v\n coda acquisizioneStrumento[ID_FALEGNAME][ID_FRESATRICE_ELETTRICA]:%d\n coda acquisizioneStrumento[ID_FALEGNAME][ID_SEGA_CIRCOLARE]:%d\n coda acquisizioneStrumento[ID_APPRENDISTA][ID_FRESATRICE_ELETTRICA]:%d\n coda acquisizioneStrumento[ID_APPRENDISTA][ID_SEGA_CIRCOLARE]:%d\n\n\n", kwUtilizzati, kwCheConsuma, prioritaFresatrici, prioritaSeghe, strumentiLiberi, len(acquisizioneStrumento[ID_FALEGNAME][ID_FRESATRICE_ELETTRICA]), len(acquisizioneStrumento[ID_FALEGNAME][ID_SEGA_CIRCOLARE]), len(acquisizioneStrumento[ID_APPRENDISTA][ID_FRESATRICE_ELETTRICA]), len(acquisizioneStrumento[ID_APPRENDISTA][ID_SEGA_CIRCOLARE]))

		/*
			da gestire
				var acquisizioneStrumento [N_TIPOLOGIE_OPERAI][N_TIPOLOGIE_STRUMENTI]Richiesta
				var rilascioStrumento chan Richiesta
		*/
		select {
		// acquisizione falegname fresatrice elettrica,controllo disponibilità fresatrice, controllo disponibilità kw,  priorità allo strumento attualmente disponibile in maggior numero
		case ric := <-when(strumentiLiberi[ID_FRESATRICE_ELETTRICA] > 0 && kwUtilizzati+EF <= MAX && (prioritaFresatrici || codaSeghe == 0), acquisizioneStrumento[ID_FALEGNAME][ID_FRESATRICE_ELETTRICA]):
			strumentiLiberi[ric.tipo]--
			kwUtilizzati += kwCheConsuma[ric.tipo]
			ric.ack <- 1

		// acquisizione falegname sega,controllo disponibilità sega, controllo disponibilità kw,  priorità allo strumento attualmente disponibile in maggior numero
		case ric := <-when(strumentiLiberi[ID_SEGA_CIRCOLARE] > 0 && kwUtilizzati+ES <= MAX && (prioritaSeghe || codaFresatrici == 0), acquisizioneStrumento[ID_FALEGNAME][ID_SEGA_CIRCOLARE]):
			strumentiLiberi[ric.tipo]--
			kwUtilizzati += kwCheConsuma[ric.tipo]
			ric.ack <- 1

		// acquisizione apprendista fresatrice elettrica,controllo disponibilità fresatrice, controllo disponibilità kw,  priorità allo strumento attualmente disponibile in maggior numero, priorità ai falegnami
		case ric := <-when(strumentiLiberi[ID_FRESATRICE_ELETTRICA] > 0 && kwUtilizzati+EF <= MAX && (prioritaFresatrici || codaSeghe == 0) && codaFresatriciFalegnami == 0, acquisizioneStrumento[ID_APPRENDISTA][ID_FRESATRICE_ELETTRICA]):
			strumentiLiberi[ric.tipo]--
			kwUtilizzati += kwCheConsuma[ric.tipo]
			ric.ack <- 1

		// acquisizione apprendista sega,controllo disponibilità sega, controllo disponibilità kw,  priorità allo strumento attualmente disponibile in maggior numero, priorità ai falegnami
		case ric := <-when(strumentiLiberi[ID_SEGA_CIRCOLARE] > 0 && kwUtilizzati+ES <= MAX && (prioritaSeghe || codaFresatrici == 0) && codaSegheFalegnami == 0, acquisizioneStrumento[ID_APPRENDISTA][ID_SEGA_CIRCOLARE]):
			strumentiLiberi[ric.tipo]--
			kwUtilizzati += kwCheConsuma[ric.tipo]
			ric.ack <- 1

		// rilascio di uno strumento
		case ric := <-rilascioStrumento:
			strumentiLiberi[ric.tipo]++
			kwUtilizzati -= kwCheConsuma[ric.tipo]
			ric.ack <- 1

		case <-endServerCmd:
			fmt.Println("SERVER: la falegnameria chiude")
			done <- true
			return
		}
	}
}

func main() {
	//init canali
	/*
		da gestire
			var acquisizioneStrumento [N_TIPOLOGIE_OPERAI][N_TIPOLOGIE_STRUMENTI]Richiesta
			var rilascioStrumento chan Richiesta
	*/
	for o := 0; o < N_TIPOLOGIE_OPERAI; o++ {
		for s := 0; s < N_TIPOLOGIE_STRUMENTI; s++ {
			acquisizioneStrumento[o][s] = make(chan Richiesta, MAXBUFF)
		}
	}
	rilascioStrumento = make(chan Richiesta, MAXBUFF)

	//avvio goroutine
	go server()

	for i := 0; i < MAXPROC; i++ {
		go operaio(i, rand.Intn(N_TIPOLOGIE_OPERAI))
	}
	for i := 0; i < MAXPROC; i++ {
		<-done
	}
	endServerCmd <- true
	<-done //attesa conferma del server

	fmt.Printf("\n MAIN: FINE PROGRAMMA\n")
}
