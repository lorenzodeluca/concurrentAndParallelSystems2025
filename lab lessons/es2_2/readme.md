## Esercizio 2.2

Variante dell’esercizio 2.1.

I visitatori possono accedere al parco **a gruppi**.  
Ogni gruppo può essere composto al massimo da **5 persone**.

Il parco ha una capacità limitata pari a **MaxP**: pertanto non può accogliere più di **MaxP persone contemporaneamente**.

Il parco è molto esteso e non può essere visitato a piedi; per la visita del parco sono disponibili **MaxA auto elettriche**, destinate al trasporto di gruppi di **1–5 persone**.

### Specifiche del comportamento

Ogni gruppo:
1. richiede l’accesso al parco alla biglietteria, acquisendo contestualmente l’auto;
2. visita il parco per un tempo arbitrario;
3. esce dal parco, restituendo l’auto usata per la visita.

### Obiettivo

Realizzare un’applicazione concorrente in **C/pthread**, nella quale **ogni gruppo sia rappresentato da un thread distinto** e la sincronizzazione venga ottenuta tramite **semafori POSIX**.

L’applicazione deve garantire che **l’ordine di ingresso al parco rispetti l’ordine cronologico di arrivo alla biglietteria**.
