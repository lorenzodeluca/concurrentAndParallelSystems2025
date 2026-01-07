# Pattern di Programmazione Concorrente in Go

---

## Indice

1. [Pattern Architetturali Principali](#1-pattern-architetturali-principali)
2. [Pattern di Priorità](#2-pattern-di-priorità)
3. [Pattern di Sincronizzazione](#3-pattern-di-sincronizzazione)
4. [Pattern di Gestione Gruppi](#4-pattern-di-gestione-gruppi)
5. [Pattern di Equivalenza Risorse](#5-pattern-di-equivalenza-risorse)
6. [Pattern di Terminazione](#6-pattern-di-terminazione)
7. [Strutture Dati Comuni](#7-strutture-dati-comuni)
8. [Checklist Anti-Deadlock](#8-checklist-anti-deadlock)
9. [Strategia di Risoluzione](#9-strategia-di-risoluzione)

---

## 1. PATTERN ARCHITETTURALI PRINCIPALI

### Pattern A: Accesso a Risorsa Limitata con Priorità

**Quando usarlo**: Più attori competono per una risorsa con capacità limitata.

**Caratteristiche**:
- Capacità massima (MAX, N, ecc.)
- Code di attesa per tipo di richiedente
- Priorità statiche o dinamiche

**Esempi**:
- Ponte (capacità MAX, priorità Nord)
- Palestra (capacità MAX, priorità trainer > utenti)
- Sala attesa (capacità MAXS, priorità amministratori > privati)

**Template di implementazione**:

```go
// Costanti
const CAPACITA_MAX = 10
const TIPO_PRIORITA_ALTA = 0
const TIPO_PRIORITA_BASSA = 1

// Canali separati per tipo
var richiestaAccessoAlta chan Richiesta
var richiestaAccessoBassa chan Richiesta

// Server
func server() {
    capacitaAttuale := 0
    
    for {
        select {
        // Priorità alta
        case r := <-when(
            capacitaAttuale < CAPACITA_MAX,
            richiestaAccessoAlta):
            capacitaAttuale++
            r.ack <- 1
            
        // Priorità bassa (solo se alta vuota)
        case r := <-when(
            capacitaAttuale < CAPACITA_MAX &&
            len(richiestaAccessoAlta) == 0,
            richiestaAccessoBassa):
            capacitaAttuale++
            r.ack <- 1
            
        // Rilascio risorsa
        case <-notificaUscita:
            capacitaAttuale--
        }
    }
}
```

---

### Pattern B: Senso Unico Alternato

**Quando usarlo**: Una risorsa (strada, ponte, corridoio) può essere usata in una sola direzione alla volta.

**Caratteristiche**:
- Due direzioni mutualmente esclusive
- Possibile capacità per direzione
- Priorità tra direzioni

**Esempi**:
- Ponte (Nord/Sud con priorità)
- Mostra fotografica (corridoio IN/OUT)
- Castello con spazzaneve (salita/discesa)

**Template di implementazione**:

```go
const CAPACITA_PONTE = 5
const DIREZIONE_NORD = 0
const DIREZIONE_SUD = 1

var richiestaAccessoNord chan Richiesta
var richiestaAccessoSud chan Richiesta
var notificaUscitaNord chan int
var notificaUscitaSud chan int

func server() {
    veicoliNord := 0
    veicoliSud := 0
    
    for {
        select {
        // Accesso Nord (priorità)
        case r := <-when(
            veicoliNord < CAPACITA_PONTE &&
            veicoliSud == 0,  // senso unico
            richiestaAccessoNord):
            veicoliNord++
            r.ack <- 1
            
        // Accesso Sud (solo se Nord vuota)
        case r := <-when(
            veicoliSud < CAPACITA_PONTE &&
            veicoliNord == 0 &&
            len(richiestaAccessoNord) == 0,  // priorità Nord
            richiestaAccessoSud):
            veicoliSud++
            r.ack <- 1
            
        // Uscita Nord
        case <-notificaUscitaNord:
            veicoliNord--
            
        // Uscita Sud
        case <-notificaUscitaSud:
            veicoliSud--
        }
    }
}
```

---

### Pattern C: Assegnazione Risorse con Supervisione

**Quando usarlo**: Utenti richiedono risorse che devono essere assegnate da un pool limitato.

**Caratteristiche**:
- Pool di risorse (operatori, trainer, commessi)
- Ogni risorsa può servire N utenti contemporaneamente
- Rilascio risorse quando utente termina

**Esempi**:
- Negozio (commessi che supervisionano 0-3 clienti)
- Palestra (trainer per lezioni private)
- Stadio (operatori per controlli sicurezza)

**Template di implementazione**:

```go
const NUM_RISORSE = 5
const MAX_UTENTI_PER_RISORSA = 3

type Risorsa struct {
    occupata              bool
    idUtentiAssegnati     [MAX_UTENTI_PER_RISORSA]int
    numeroUtentiAssegnati int
    inAttesaDiUscire      bool
    ackUscita             chan bool
}

func server() {
    risorse := make([]Risorsa, NUM_RISORSE)
    risorseDisponibili := NUM_RISORSE  // con slot liberi
    
    // Inizializzazione
    for i := 0; i < NUM_RISORSE; i++ {
        risorse[i].numeroUtentiAssegnati = 0
        for j := 0; j < MAX_UTENTI_PER_RISORSA; j++ {
            risorse[i].idUtentiAssegnati[j] = -1
        }
    }
    
    for {
        select {
        // Richiesta utente
        case r := <-when(risorseDisponibili > 0, richiestaUtente):
            // Cerca risorsa con slot libero
            trovato := false
            for idRisorsa := 0; idRisorsa < NUM_RISORSE && !trovato; idRisorsa++ {
                if risorse[idRisorsa].numeroUtentiAssegnati < MAX_UTENTI_PER_RISORSA {
                    // Trova slot libero
                    for slot := 0; slot < MAX_UTENTI_PER_RISORSA && !trovato; slot++ {
                        if risorse[idRisorsa].idUtentiAssegnati[slot] == -1 {
                            risorse[idRisorsa].idUtentiAssegnati[slot] = r.id
                            risorse[idRisorsa].numeroUtentiAssegnati++
                            
                            if risorse[idRisorsa].numeroUtentiAssegnati == MAX_UTENTI_PER_RISORSA {
                                risorseDisponibili--
                            }
                            
                            trovato = true
                            r.ack <- idRisorsa
                        }
                    }
                }
            }
            
        // Uscita utente
        case r := <-notificaUscitaUtente:
            // Trova e libera lo slot
            for idRisorsa := 0; idRisorsa < NUM_RISORSE; idRisorsa++ {
                for slot := 0; slot < MAX_UTENTI_PER_RISORSA; slot++ {
                    if risorse[idRisorsa].idUtentiAssegnati[slot] == r.id {
                        risorse[idRisorsa].idUtentiAssegnati[slot] = -1
                        
                        if risorse[idRisorsa].numeroUtentiAssegnati == MAX_UTENTI_PER_RISORSA {
                            risorseDisponibili++
                        }
                        risorse[idRisorsa].numeroUtentiAssegnati--
                        
                        // Se la risorsa era in attesa di uscire
                        if risorse[idRisorsa].inAttesaDiUscire && 
                           risorse[idRisorsa].numeroUtentiAssegnati == 0 {
                            risorse[idRisorsa].ackUscita <- true
                            risorse[idRisorsa].inAttesaDiUscire = false
                            risorseDisponibili--
                        }
                        break
                    }
                }
            }
        }
    }
}
```

---

### Pattern D: Deposito/Magazzino con Fornitori e Consumatori

**Quando usarlo**: Magazzino con scorte limitate, fornitori che riforniscono, consumatori che prelevano.

**Caratteristiche**:
- Capacità massima per tipo di merce
- Fornitori riempiono (o consegnano lotti)
- Consumatori prelevano (lotti o unità)
- Priorità tra fornitori/consumatori
- Esclusione mutua tra fornitori e consumatori

**Esempi**:
- Stabilimento auto (deposito cerchi/pneumatici)
- Magazzino mascherine (scaffali FFP2/chirurgiche)
- Negozio (mascherine al punto distribuzione)

**Template di implementazione**:

```go
const CAPACITA_DEPOSITO = 20
const DIMENSIONE_LOTTO_CONSEGNA = 10
const DIMENSIONE_LOTTO_PRELIEVO = 3

var richiestaConsegna chan Richiesta
var notificaFineConsegna chan Richiesta
var richiestaPrelievo chan Richiesta
var notificaFinePrelievo chan Richiesta

func deposito() {
    scorteDisponibili := 0
    fornitoriInConsegna := 0
    consumatoriInPrelievo := 0
    
    for {
        select {
        // Consegna fornitore
        case r := <-when(
            scorteDisponibili + DIMENSIONE_LOTTO_CONSEGNA <= CAPACITA_DEPOSITO &&
            consumatoriInPrelievo == 0,  // esclusione mutua
            richiestaConsegna):
            
            fornitoriInConsegna++
            r.ack <- 1
            
        // Fine consegna
        case r := <-notificaFineConsegna:
            scorteDisponibili += DIMENSIONE_LOTTO_CONSEGNA
            fornitoriInConsegna--
            r.ack <- 1
            
        // Prelievo consumatore
        case r := <-when(
            scorteDisponibili >= DIMENSIONE_LOTTO_PRELIEVO &&
            fornitoriInConsegna == 0,  // esclusione mutua
            richiestaPrelievo):
            
            consumatoriInPrelievo++
            scorteDisponibili -= DIMENSIONE_LOTTO_PRELIEVO
            r.ack <- 1
            
        // Fine prelievo
        case r := <-notificaFinePrelievo:
            consumatoriInPrelievo--
            r.ack <- 1
        }
    }
}
```

---

## 2. PATTERN DI PRIORITÀ

### A. Priorità Statica con len()

**Quando usarlo**: Un tipo ha sempre priorità su un altro.

**Implementazione**:

```go
select {
    // Priorità alta (servito sempre se presente)
    case r := <-richiestaAlta:
        // serve richiesta alta
        
    // Priorità bassa (servito solo se alta vuota)
    case r := <-when(len(richiestaAlta) == 0, richiestaMedia):
        // serve richiesta media
        
    // Priorità bassissima (servito solo se alta e media vuote)
    case r := <-when(
        len(richiestaAlta) == 0 && 
        len(richiestaMedia) == 0, 
        richiestaMedia):
        // serve richiesta bassa
}
```

**Esempio con 3 livelli (Palestra)**:
```go
// Priorità: trainer > area corsi > area pesi
select {
    case r := <-richiestaIngressoTrainer:  // priorità 1
        // ...
        
    case r := <-when(
        len(richiestaIngressoTrainer) == 0,
        richiestaIngressoAreaCorsi):  // priorità 2
        // ...
        
    case r := <-when(
        len(richiestaIngressoTrainer) == 0 &&
        len(richiestaIngressoAreaCorsi) == 0,
        richiestaIngressoAreaPesi):  // priorità 3
        // ...
}
```

---

### B. Priorità Dinamica Basata su Contatori

**Quando usarlo**: La priorità cambia in base allo stato del sistema (es. tribuna più popolata).

**⚠️ ATTENZIONE DEADLOCK**: Servire ANCHE quando la coda prioritaria è vuota!

**Implementazione**:

```go
// Priorità dinamica: favorisce chi ha meno
prioritaA := contatoreA <= contatoreB

select {
    // Serve A quando ha priorità
    case r := <-when(prioritaA, richiestaA):
        contatoreA++
        // ...
        
    // Serve B quando ha priorità
    case r := <-when(!prioritaA, richiestaB):
        contatoreB++
        // ...
        
    // ⚠️ IMPORTANTE: Serve A anche quando NON ha priorità ma B è vuota
    case r := <-when(
        !prioritaA && len(richiestaB) == 0,
        richiestaA):
        contatoreA++
        // ...
        
    // ⚠️ IMPORTANTE: Serve B anche quando NON ha priorità ma A è vuota
    case r := <-when(
        prioritaA && len(richiestaA) == 0,
        richiestaB):
        contatoreB++
        // ...
}
```

**Esempio reale (Stadio - tribuna più popolata)**:
```go
spettatoriInTribunaLocali := 0
spettatoriInTribunaOspiti := 0

prioritaLocali := spettatoriInTribunaLocali >= spettatoriInTribunaOspiti

select {
    case r := <-when(prioritaLocali, richiestaControlloLocali):
        // serve Locali con priorità
        
    case r := <-when(!prioritaLocali, richiestaControlloOspiti):
        // serve Ospiti con priorità
        
    case r := <-when(
        prioritaLocali && len(richiestaControlloLocali) == 0,
        richiestaControlloOspiti):
        // serve Ospiti anche se Locali ha priorità (evita deadlock)
        
    case r := <-when(
        !prioritaLocali && len(richiestaControlloOspiti) == 0,
        richiestaControlloLocali):
        // serve Locali anche se Ospiti ha priorità (evita deadlock)
}
```

---

### C. Priorità a Cascata (3+ livelli)

**Quando usarlo**: Più di 2 livelli di priorità.

**Implementazione**:

```go
select {
    // Livello 1: massima priorità
    case r := <-canale1:
        // ...
        
    // Livello 2: priorità media (solo se livello 1 vuoto)
    case r := <-when(len(canale1) == 0, canale2):
        // ...
        
    // Livello 3: priorità bassa (solo se livelli 1 e 2 vuoti)
    case r := <-when(
        len(canale1) == 0 && len(canale2) == 0,
        canale3):
        // ...
        
    // Livello 4: minima priorità (solo se tutti gli altri vuoti)
    case r := <-when(
        len(canale1) == 0 && 
        len(canale2) == 0 && 
        len(canale3) == 0,
        canale4):
        // ...
}
```

---

## 3. PATTERN DI SINCRONIZZAZIONE

### A. Operazione in Più Fasi

**Quando usarlo**: Un'operazione richiede più interazioni con il server.

**Esempi**:
- Stadio: acquisto biglietto → controllo → ingresso tribuna
- Mostra: ingresso corridoio → sala → uscita corridoio
- Negozio: ingresso → shopping → uscita

**Template goroutine utente**:

```go
func utente(id int) {
    r := Richiesta{id: id, ack: make(chan int, MAXBUFF)}
    
    // FASE 1: Richiesta iniziale
    richiestaFase1 <- r
    risultato1 := <-r.ack
    fmt.Printf("[UTENTE %d] Completata fase 1\n", id)
    
    // FASE 2: Operazione (tempo non trascurabile)
    sleepRandTime(TEMPO_FASE2)
    
    richiestaFase2 <- r
    <-r.ack
    fmt.Printf("[UTENTE %d] Completata fase 2\n", id)
    
    // FASE 3: Completamento
    sleepRandTime(TEMPO_FASE3)
    
    notificaCompletamento <- r
    <-r.ack
    fmt.Printf("[UTENTE %d] Operazione completata\n", id)
    
    done <- true
}
```

---

### B. Attesa Condizionale per Uscita

**Quando usarlo**: Una risorsa vuole uscire ma deve attendere che una condizione si verifichi.

**Esempi**:
- Commesso che vuole uscire ma ha ancora clienti assegnati
- Personal trainer che vuole uscire ma sta dando una lezione

**Template di implementazione**:

```go
type Risorsa struct {
    inAttesaDiUscire    bool
    condizioneBloccante bool  // es: ha utenti assegnati
    ackUscita           chan bool
}

// Nel server
select {
    // Richiesta di uscita
    case r := <-richiestaUscita:
        if risorse[r.id].condizioneBloccante {
            // Non può uscire ora, mette in attesa
            risorse[r.id].inAttesaDiUscire = true
            risorse[r.id].ackUscita = r.ack
            fmt.Printf("Risorsa %d in attesa di uscire\n", r.id)
        } else {
            // Può uscire subito
            risorse[r.id].inAttesaDiUscire = false
            r.ack <- true
            fmt.Printf("Risorsa %d esce\n", r.id)
        }
        
    // Quando la condizione si sblocca (es: utente esce)
    case r := <-notificaUscitaUtente:
        // ... libera utente da risorsa ...
        
        // Verifica se la risorsa era in attesa di uscire
        if risorse[idRisorsa].inAttesaDiUscire && 
           !risorse[idRisorsa].condizioneBloccante {
            // Ora può uscire
            risorse[idRisorsa].ackUscita <- true
            risorse[idRisorsa].inAttesaDiUscire = false
            fmt.Printf("Risorsa %d può finalmente uscire\n", idRisorsa)
        }
}
```

---

### C. Goroutine Separate per Operazioni Lunghe

**Quando usarlo**: Per simulare operazioni non trascurabili senza bloccare il server.

**Esempi**:
- Controllo di sicurezza al varco (tempo non trascurabile)
- Rifornimento scaffale da parte del fornitore

**⚠️ IMPORTANTE**: Il server autorizza subito, poi l'operazione continua in background.

**Template di implementazione**:

```go
const NUM_OPERATORI = 5

type Operatore struct {
    occupato          bool
    idUtenteAssegnato int
}

var operatori []Operatore
var operatoriLiberi int

select {
    // Richiesta di operazione lunga
    case r := <-when(operatoriLiberi > 0, richiestaOperazione):
        // Trova operatore libero
        for idOp := 0; idOp < NUM_OPERATORI; idOp++ {
            if !operatori[idOp].occupato {
                operatori[idOp].occupato = true
                operatori[idOp].idUtenteAssegnato = r.id
                operatoriLiberi--
                
                // Autorizzazione immediata
                r.ack <- idOp
                
                // Lancia goroutine separata per l'operazione
                go func(operatoreID int, richiesta Richiesta) {
                    fmt.Printf("Operatore %d inizia operazione\n", operatoreID)
                    
                    // Operazione lunga
                    sleepRandTime(TEMPO_OPERAZIONE)
                    
                    // Notifica completamento
                    notificaCompletamento <- richiesta
                }(idOp, r)
                
                break
            }
        }
        
    // Completamento operazione
    case r := <-notificaCompletamento:
        // Trova e libera operatore
        for idOp := 0; idOp < NUM_OPERATORI; idOp++ {
            if operatori[idOp].idUtenteAssegnato == r.id {
                operatori[idOp].occupato = false
                operatori[idOp].idUtenteAssegnato = -1
                operatoriLiberi++
                
                r.ack <- 1
                fmt.Printf("Operatore %d liberato\n", idOp)
                break
            }
        }
}
```

---

## 4. PATTERN DI GESTIONE GRUPPI

### A. Gruppo come Unità Atomica

**Quando usarlo**: Un gruppo (scolaresca, privato+accompagnatore) viene trattato come singola entità.

**Caratteristiche**:
- Il gruppo entra/esce tutto insieme
- Occupa N posti (es: scolaresca = 25 persone)
- Viene conteggiato come singola unità in alcuni contatori

**Esempi**:
- Scolaresca di 25 persone nella mostra
- Privato con accompagnatore (2 posti)

**Template di implementazione**:

```go
const DIMENSIONE_GRUPPO = 25
const CAPACITA_MAX = 40

gruppiPresenti := 0
personePresenti := 0

select {
    // Ingresso gruppo
    case r := <-when(
        personePresenti + DIMENSIONE_GRUPPO <= CAPACITA_MAX,
        richiestaIngressoGruppo):
        
        gruppiPresenti++
        personePresenti += DIMENSIONE_GRUPPO
        r.ack <- 1
        
    // Uscita gruppo
    case r := <-notificaUscitaGruppo:
        gruppiPresenti--
        personePresenti -= DIMENSIONE_GRUPPO
        r.ack <- 1
}
```

**Esempio con accompagnatore (2 posti)**:

```go
const CAPACITA_SALA = 10

select {
    // Singolo (1 posto)
    case r := <-when(
        personeInSala < CAPACITA_SALA,
        richiestaIngressoSingolo):
        personeInSala++
        r.ack <- 1
        
    // Con accompagnatore (2 posti)
    case r := <-when(
        personeInSala + 2 <= CAPACITA_SALA,
        richiestaIngressoConAccompagnatore):
        personeInSala += 2
        r.ack <- 1
}
```

---

### B. Gruppo Blocca Direzione Opposta

**Quando usarlo**: La presenza di un gruppo in una direzione impedisce accesso in direzione opposta.

**Esempio**: Scolaresca nel corridoio blocca transito opposto.

**Template di implementazione**:

```go
const DIREZIONE_IN = 0
const DIREZIONE_OUT = 1

gruppiInDirezione := [2]int{0, 0}
personeInDirezione := [2]int{0, 0}

select {
    // Gruppo in direzione IN
    case r := <-when(
        personeInDirezione[DIREZIONE_OUT] == 0 &&  // nessuno in direzione opposta
        personeInDirezione[DIREZIONE_IN] + DIMENSIONE_GRUPPO <= CAPACITA,
        richiestaGruppoIN):
        
        gruppiInDirezione[DIREZIONE_IN]++
        personeInDirezione[DIREZIONE_IN] += DIMENSIONE_GRUPPO
        r.ack <- 1
        
    // Singolo in direzione IN (solo se nessun gruppo in OUT)
    case r := <-when(
        gruppiInDirezione[DIREZIONE_OUT] == 0 &&  // gruppi bloccano tutto
        personeInDirezione[DIREZIONE_IN] < CAPACITA,
        richiestaIngressoSingoloIN):
        
        personeInDirezione[DIREZIONE_IN]++
        r.ack <- 1
}
```

---

## 5. PATTERN DI EQUIVALENZA RISORSE

### A. Conversione con Peso

**Quando usarlo**: Diverse entità hanno "peso" diverso per la stessa capacità.

**Esempio**: Pedoni = 1 persona, Auto = 10 persone equivalenti.

**Template di implementazione**:

```go
const CAPACITA_PONTE = 35
const PESO_PEDONE = 1
const PESO_AUTO = 10

personeEquivalenti := 0
numPedoni := 0
numAuto := 0

select {
    // Ingresso pedone
    case r := <-when(
        personeEquivalenti + PESO_PEDONE <= CAPACITA_PONTE,
        richiestaIngressoPedone):
        
        personeEquivalenti += PESO_PEDONE
        numPedoni++
        r.ack <- 1
        
    // Ingresso auto
    case r := <-when(
        personeEquivalenti + PESO_AUTO <= CAPACITA_PONTE,
        richiestaIngressoAuto):
        
        personeEquivalenti += PESO_AUTO
        numAuto++
        r.ack <- 1
        
    // Uscita (deve sapere il tipo per decrementare correttamente)
    case r := <-notificaUscita:
        if r.tipo == TIPO_PEDONE {
            personeEquivalenti -= PESO_PEDONE
            numPedoni--
        } else {
            personeEquivalenti -= PESO_AUTO
            numAuto--
        }
        r.ack <- 1
}
```

---

### B. Slot Multipli con Fallback

**Quando usarlo**: Una risorsa può usare slot di tipo diverso (con preferenza).

**Esempio**: Automobile può usare parcheggio standard o maxi (se standard pieno).

**Template di implementazione**:

```go
const CAPACITA_STANDARD = 10
const CAPACITA_MAXI = 5

slotStandard := CAPACITA_STANDARD
slotMaxi := CAPACITA_MAXI

// Tipo di parcheggio
const TIPO_STANDARD = 0
const TIPO_MAXI = 1

select {
    // Camper (solo maxi)
    case r := <-when(slotMaxi > 0, richiestaIngressoCamper):
        slotMaxi--
        r.ack <- TIPO_MAXI
        
    // Automobile (preferisce standard, fallback su maxi)
    case r := <-when(
        slotStandard > 0 || slotMaxi > 0,
        richiestaIngressoAuto):
        
        var tipoAssegnato int
        if slotStandard > 0 {
            slotStandard--
            tipoAssegnato = TIPO_STANDARD
        } else {
            slotMaxi--
            tipoAssegnato = TIPO_MAXI
        }
        r.ack <- tipoAssegnato
        
    // Uscita (deve sapere quale tipo rilasciare)
    case r := <-notificaUscita:
        if r.tipoParcheggio == TIPO_STANDARD {
            slotStandard++
        } else {
            slotMaxi++
        }
        r.ack <- 1
}
```

---

## 6. PATTERN DI TERMINAZIONE

### A. Terminazione a Cascata

**Quando usarlo**: Sempre! Ordine corretto di terminazione.

**Ordine standard**:
1. Utenti/Clienti (aspetta che tutti terminino)
2. Risorse/Fornitori (invia segnale, aspetta conferma)
3. Server (invia segnale, aspetta conferma)

**Template di implementazione**:

```go
func main() {
    // Avvio goroutine
    go server()
    
    for i := 0; i < NUM_FORNITORI; i++ {
        go fornitore(i)
    }
    
    for i := 0; i < NUM_UTENTI; i++ {
        go utente(i)
    }
    
    // === FASE 1: Attendi terminazione utenti ===
    for i := 0; i < NUM_UTENTI; i++ {
        <-done
    }
    fmt.Printf("[MAIN] Tutti gli utenti hanno terminato\n")
    
    // === FASE 2: Termina fornitori/risorse ===
    for i := 0; i < NUM_FORNITORI; i++ {
        terminaFornitore <- true
    }
    for i := 0; i < NUM_FORNITORI; i++ {
        <-done
    }
    fmt.Printf("[MAIN] Tutti i fornitori hanno terminato\n")
    
    // === FASE 3: Termina server ===
    terminaServer <- true
    <-done
    fmt.Printf("[MAIN] Server terminato\n")
    
    fmt.Printf("[MAIN] Applicazione terminata\n")
}
```

---

### B. Terminazione con Rifiuto Richieste

**Quando usarlo**: Quando il server deve rifiutare nuove richieste durante la chiusura.

**Esempi**:
- Fornitore che riceve -1 come ACK per terminare
- Robot che riceve -1 e sa che deve smettere

**Template di implementazione**:

```go
func server() {
    terminazioneProgrammata := false
    
    for {
        select {
        // Operazioni normali
        case r := <-when(!terminazioneProgrammata, richiestaOperazione):
            // processa richiesta normalmente
            r.ack <- 1
            
        // Segnale di inizio terminazione
        case <-termina:
            terminazioneProgrammata = true
            fmt.Printf("[SERVER] Inizio procedura di chiusura\n")
            
        // Rifiuta nuove richieste dopo terminazione
        case r := <-when(terminazioneProgrammata, richiestaOperazione):
            r.ack <- -1  // segnale di terminazione
            
        // Terminazione finale
        case <-chiudi:
            fmt.Printf("[SERVER] Chiusura definitiva\n")
            done <- true
            return
        }
    }
}
```

**Goroutine che riceve il segnale di terminazione**:

```go
func fornitore(id int) {
    for {
        richiestaOperazione <- r
        risultato := <-r.ack
        
        if risultato == -1 {
            fmt.Printf("[FORNITORE %d] Ricevuto segnale di terminazione\n", id)
            done <- true
            return
        }
        
        // continua operazioni...
    }
}
```

---

## 7. STRUTTURE DATI COMUNI

### A. Array di Canali per Tipo

**Quando usarlo**: Diverse tipologie di richiedenti (sempre!).

```go
const NUM_TIPI = 3
const TIPO_A = 0
const TIPO_B = 1
const TIPO_C = 2

var richiestaPerTipo [NUM_TIPI]chan Richiesta

// Inizializzazione nel main
func main() {
    for i := 0; i < NUM_TIPI; i++ {
        richiestaPerTipo[i] = make(chan Richiesta, MAXBUFF)
    }
    
    // Uso nel codice
    richiestaPerTipo[tipoUtente] <- r
}
```

---

### B. Array/Slice di Struct per Stato Risorse

**Quando usarlo**: Pool di risorse con stato (operatori, trainer, commessi, ecc.).

```go
const NUM_RISORSE = 5

type Risorsa struct {
    occupata              bool
    idUtenteAssegnato     int
    inAttesaDiUscire      bool
    ackUscita             chan bool
}

func server() {
    risorse := make([]Risorsa, NUM_RISORSE)
    
    // Inizializzazione
    for i := 0; i < NUM_RISORSE; i++ {
        risorse[i].occupata = false
        risorse[i].idUtenteAssegnato = -1
        risorse[i].inAttesaDiUscire = false
        risorse[i].ackUscita = nil
    }
    
    // Uso
    for i := 0; i < NUM_RISORSE; i++ {
        if !risorse[i].occupata {
            risorse[i].occupata = true
            risorse[i].idUtenteAssegnato = r.id
            // ...
        }
    }
}
```

---

### C. Mappa per Associazioni

**Quando usarlo**: Associare informazioni a un ID (biglietti, prenotazioni, ecc.).

```go
type Info struct {
    id              int
    tribunaAssegnata int
    // ... altri campi
}

var mappaInfo = make(map[int]Info)

// Inserimento
mappaInfo[id] = Info{
    id:              id,
    tribunaAssegnata: tribuna,
}

// Recupero
info := mappaInfo[id]
fmt.Printf("Tribuna: %s\n", labelTribuna[info.tribunaAssegnata])

// Ricerca per valore
for _, info := range mappaInfo {
    if info.id == idCercato {
        // trovato
    }
}
```

---

### D. Struct Richiesta Standard

**Sempre usare questa struttura**:

```go
type Richiesta struct {
    id   int      // identificativo richiedente
    tipo int      // tipo di richiesta (opzionale)
    ack  chan int // canale per conferme/risultati
}

// Uso nella goroutine
func utente(id int) {
    r := Richiesta{
        id:   id,
        tipo: TIPO_A,
        ack:  make(chan int, MAXBUFF),
    }
    
    richiestaOperazione <- r
    risultato := <-r.ack
}
```

---

## 8. CHECKLIST ANTI-DEADLOCK

### ⚠️ Errori Comuni che Causano Deadlock

#### ❌ ERRORE 1: Priorità dinamica senza fallback
```go
// ✗ SBAGLIATO - CAUSA DEADLOCK
prioritaA := contatoreA >= contatoreB

select {
    case r := <-when(prioritaA, richiestaA):
        // serve A
    case r := <-when(!prioritaA, richiestaB):
        // serve B
}
// PROBLEMA: se prioritaA è true ma richiestaA è vuota, 
// nessuno viene servito anche se richiestaB ha elementi!
```

```go
// ✓ CORRETTO
prioritaA := contatoreA >= contatoreB

select {
    case r := <-when(prioritaA, richiestaA):
        // serve A con priorità
    case r := <-when(!prioritaA, richiestaB):
        // serve B con priorità
    case r := <-when(prioritaA && len(richiestaA) == 0, richiestaB):
        // FALLBACK: serve B anche se A ha priorità ma coda A vuota
    case r := <-when(!prioritaA && len(richiestaB) == 0, richiestaA):
        // FALLBACK: serve A anche se B ha priorità ma coda B vuota
}
```

---

#### ❌ ERRORE 2: Esclusione mutua troppo restrittiva
```go
// ✗ SBAGLIATO - STARVATION
select {
    case r := <-when(
        contA == 0 && contB == 0,  // troppo restrittivo
        richiestaA):
        contA++
}
// PROBLEMA: se c'è sempre qualcuno, nessuno può entrare mai
```

```go
// ✓ CORRETTO - Esclusione mutua con condizioni ragionevoli
select {
    case r := <-when(
        contB == 0,  // solo direzione opposta
        richiestaA):
        contA++
}
```

---

#### ❌ ERRORE 3: Dimenticare di liberare risorse
```go
// ✗ SBAGLIATO
case r := <-richiestaUscita:
    // libera utente ma dimentica di liberare operatore
    utentiPresenti--
    r.ack <- 1
```

```go
// ✓ CORRETTO
case r := <-richiestaUscita:
    utentiPresenti--
    
    // Trova e libera l'operatore associato
    for i := 0; i < NUM_OPERATORI; i++ {
        if operatori[i].idUtenteAssegnato == r.id {
            operatori[i].occupato = false
            operatori[i].idUtenteAssegnato = -1
            operatoriLiberi++
            break
        }
    }
    r.ack <- 1
```

---

### ✅ Checklist di Verifica

Prima di considerare completa la soluzione, verifica:

- [ ] **Ogni canale con priorità ha un case di fallback**
  - Se priorità dinamica → 4 case (2 con priorità + 2 fallback)
  - Se priorità statica → case con `len() == 0`

- [ ] **Esclusione mutua permette progressione**
  - Controlla che non ci siano condizioni tipo `A==0 && B==0 && C==0`
  - Se senso unico, basta controllare direzione opposta

- [ ] **Risorse vengono sempre liberate**
  - Ogni operatore/risorsa assegnata viene liberata all'uscita
  - Contatori decrementati correttamente

- [ ] **Operazioni lunghe non bloccano il server**
  - Usa goroutine separate per operazioni con `sleepRandTime`

- [ ] **Attese condizionali vengono svegliate**
  - Se una risorsa è in attesa, controlla quando si sblocca

- [ ] **Terminazione a cascata corretta**
  - Ordine: utenti → fornitori/risorse → server

- [ ] **Capacità e vincoli rispettati**
  - `<` vs `<=` nelle condizioni
  - Gruppi conteggiati correttamente (×25, ×10, ecc.)

---

## 9. STRATEGIA DI RISOLUZIONE

### Passo 1: Analisi del Testo

**Domande da porsi**:

1. **Quale pattern architetturale?**
   - Risorsa limitata? → Pattern A
   - Senso unico? → Pattern B
   - Assegnazione risorse? → Pattern C
   - Deposito/magazzino? → Pattern D

2. **Quanti tipi di attori?**
   - Conta i tipi di goroutine necessarie
   - Identifica chi sono (utenti, fornitori, risorse)

3. **Quali sono le priorità?**
   - Statiche o dinamiche?
   - Quanti livelli?

4. **Quali sono i vincoli?**
   - Capacità massime
   - Esclusioni mutue
   - Operazioni in più fasi

---

### Passo 2: Disegna lo Schema

```
ATTORI:
- Utente (N goroutine)
- Fornitore (M goroutine)
- Server (1 goroutine)

CANALI:
utente → server: richiestaIngresso
utente → server: notificaUscita
server → utente: ack

fornitore → server: richiestaConsegna
server → fornitore: ack

STATI SERVER:
- capacitaAttuale: int
- operatoriLiberi: int
- operatori: []Operatore
```

---

### Passo 3: Implementa per Componenti

#### 1. Costanti e Tipi
```go
const CAPACITA_MAX = 10
const TIPO_A = 0
const TIPO_B = 1

type Richiesta struct {
    id   int
    tipo int
    ack  chan int
}
```

#### 2. Canali globali
```go
var richiestaIngresso [NUM_TIPI]chan Richiesta
var notificaUscita chan Richiesta
var done chan bool
var termina chan bool
```

#### 3. Goroutine utente
```go
func utente(id int, tipo int) {
    r := Richiesta{id: id, tipo: tipo, ack: make(chan int, MAXBUFF)}
    
    // Operazioni...
    
    done <- true
}
```

#### 4. Server (select)
```go
func server() {
    // Stati
    capacitaAttuale := 0
    
    for {
        select {
        case r := <-when(..., richiestaIngresso[TIPO_A]):
            // ...
        case <-termina:
            done <- true
            return
        }
    }
}
```

#### 5. Main
```go
func main() {
    // Inizializzazione canali
    for i := 0; i < NUM_TIPI; i++ {
        richiestaIngresso[i] = make(chan Richiesta, MAXBUFF)
    }
    
    // Avvio goroutine
    go server()
    for i := 0; i < NUM_UTENTI; i++ {
        go utente(i, rand.Intn(NUM_TIPI))
    }
    
    // Terminazione a cascata
    for i := 0; i < NUM_UTENTI; i++ {
        <-done
    }
    termina <- true
    <-done
}
```

---

### Passo 4: Verifica Anti-Deadlock

Usa la **checklist** della sezione 8 per verificare:
- [ ] Priorità con fallback
- [ ] Esclusione mutua ragionevole
- [ ] Risorse liberate correttamente
- [ ] Operazioni lunghe non bloccanti

---

### Passo 5: Test Mentale

Simula questi scenari:

1. **Arrivo massivo tipo A**: Tipo B viene mai servito?
2. **Capacità piena**: Cosa succede quando si libera un posto?
3. **Risorsa in attesa**: Viene svegliata quando può uscire?
4. **Terminazione**: Tutti i goroutine terminano correttamente?

---

## TEMPLATE COMPLETO DI BASE

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

// ========== COSTANTI ==========
const (
    MAXBUFF      = 100
    NUM_UTENTI   = 20
    CAPACITA_MAX = 10
    
    TIPO_A = 0
    TIPO_B = 1
    NUM_TIPI = 2
)

// ========== STRUTTURE ==========
type Richiesta struct {
    id   int
    tipo int
    ack  chan int
}

// ========== CANALI GLOBALI ==========
var (
    richiestaIngresso [NUM_TIPI]chan Richiesta
    notificaUscita    chan Richiesta
    done              chan bool
    termina           chan bool
)

// ========== FUNZIONI UTILITY ==========
func when(condizione bool, canale chan Richiesta) chan Richiesta {
    if !condizione {
        return nil
    }
    return canale
}

func sleepRandTime(limitSecondi int) {
    if limitSecondi > 0 {
        time.Sleep(time.Duration(rand.Intn(limitSecondi)+1) * time.Second)
    }
}

// ========== GOROUTINE UTENTE ==========
func utente(id int, tipo int) {
    r := Richiesta{
        id:   id,
        tipo: tipo,
        ack:  make(chan int, MAXBUFF),
    }
    
    sleepRandTime(3)
    
    // FASE 1: Ingresso
    fmt.Printf("[UTENTE %d] Richiedo ingresso\n", id)
    richiestaIngresso[tipo] <- r
    <-r.ack
    fmt.Printf("[UTENTE %d] Entrato\n", id)
    
    // FASE 2: Operazione
    sleepRandTime(5)
    
    // FASE 3: Uscita
    fmt.Printf("[UTENTE %d] Esco\n", id)
    notificaUscita <- r
    <-r.ack
    
    done <- true
}

// ========== SERVER ==========
func server() {
    capacitaAttuale := 0
    
    fmt.Printf("[SERVER] Avvio\n")
    
    for {
        select {
        // Ingresso tipo A (priorità alta)
        case r := <-when(
            capacitaAttuale < CAPACITA_MAX,
            richiestaIngresso[TIPO_A]):
            
            capacitaAttuale++
            fmt.Printf("[SERVER] Utente %d tipo A entra (capacità: %d/%d)\n",
                r.id, capacitaAttuale, CAPACITA_MAX)
            r.ack <- 1
            
        // Ingresso tipo B (priorità bassa)
        case r := <-when(
            capacitaAttuale < CAPACITA_MAX &&
            len(richiestaIngresso[TIPO_A]) == 0,
            richiestaIngresso[TIPO_B]):
            
            capacitaAttuale++
            fmt.Printf("[SERVER] Utente %d tipo B entra (capacità: %d/%d)\n",
                r.id, capacitaAttuale, CAPACITA_MAX)
            r.ack <- 1
            
        // Uscita
        case r := <-notificaUscita:
            capacitaAttuale--
            fmt.Printf("[SERVER] Utente %d esce (capacità: %d/%d)\n",
                r.id, capacitaAttuale, CAPACITA_MAX)
            r.ack <- 1
            
        // Terminazione
        case <-termina:
            fmt.Printf("[SERVER] Termino\n")
            done <- true
            return
        }
    }
}

// ========== MAIN ==========
func main() {
    rand.Seed(time.Now().UnixNano())
    
    // Inizializzazione canali
    for i := 0; i < NUM_TIPI; i++ {
        richiestaIngresso[i] = make(chan Richiesta, MAXBUFF)
    }
    notificaUscita = make(chan Richiesta, MAXBUFF)
    done = make(chan bool)
    termina = make(chan bool)
    
    fmt.Printf("[MAIN] Avvio sistema\n")
    
    // Avvio server
    go server()
    
    // Avvio utenti
    for i := 0; i < NUM_UTENTI; i++ {
        go utente(i, rand.Intn(NUM_TIPI))
    }
    
    // Attesa terminazione utenti
    for i := 0; i < NUM_UTENTI; i++ {
        <-done
    }
    
    // Terminazione server
    termina <- true
    <-done
    
    fmt.Printf("[MAIN] Sistema terminato\n")
}
```

---

## SUGGERIMENTI FINALI PER L'ESAME

### Prima di iniziare a scrivere:
1. **Leggi tutto il testo** 2 volte
2. **Sottolinea** vincoli e priorità
3. **Identifica il pattern** architetturale
4. **Disegna lo schema** su carta
5. **Elenca le goroutine** necessarie
6. **Elenca i canali** necessari

### Durante la scrittura:
1. **Copia il template** di base
2. **Adatta le costanti** al problema
3. **Implementa il server** con i case necessari
4. **Aggiungi i fallback** per le priorità
5. **Verifica la checklist** anti-deadlock

### Debugging mentale:
- "Se arrivano 10 utenti tipo A, tipo B viene servito?"
- "Se la capacità è piena, cosa succede all'uscita?"
- "Tutte le goroutine terminano?"

### Gestione del tempo:
- **10 min**: Analisi e schema
- **30 min**: Implementazione base
- **10 min**: Verifica e test mentale
- **10 min**: Commenti e pulizia codice

---

**Buona fortuna all'esame! 🚀**
