# Esercizio 2.1

Si consideri un parco a tema al quale i visitatori possono accedere singolarmente.  
Il parco ha una capacità limitata pari a **MaxP**: pertanto non può accogliere più di **MaxP** persone contemporaneamente.

Il parco è molto esteso e non può essere visitato a piedi; pertanto per la visita del parco ogni visitatore utilizzerà un veicolo messo a disposizione dal parco.  
In particolare vengono offerti 2 tipi di veicoli:
- **biciclette** (numero totale = **MaxB**)  
- **monopattini elettrici** (numero totale = **MaxM**)

Ogni visitatore:
1. richiede l’accesso al parco alla biglietteria, acquisendo contestualmente un veicolo a sua scelta (bici o monopattino);
2. visita il parco per un tempo arbitrario;
3. esce dal parco, restituendo il veicolo usato per la visita.

---

## Obiettivo
Realizzare un’applicazione concorrente in **C/pthread**, nella quale ogni visitatore sia rappresentato da un thread distinto e la sincronizzazione venga ottenuta tramite **semafori POSIX**.
