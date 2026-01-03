with Ada.Text_IO, Ada.Integer_Text_IO;
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure PiscinaPubblica is
	MAX : constant Integer := 6; -- capacità piscina
	MAX_CHIAVI : constant Integer := 20;

	type UtenteT is range 1..10;
	type ChiaveT is range 1..MAX_CHIAVI;
	type CatUtente is (Studente, Ordinario);
	type TipoTariffa is (Abbonato, NonAbbonato);

	-- BIGLIETTERIA
	task Biglietteria is
		entry acquistaBiglietto(tariffa: in TipoTariffa; chiave: out ChiaveT);
		entry restituisciChiave(chiave: in ChiaveT);
	end;
	task body Biglietteria is
		chiavi : array(ChiaveT) of Boolean;
		libere : Integer := MAX_CHIAVI;

		function GetChiave return ChiaveT is
			c : ChiaveT;
		begin
			for I in ChiaveT loop
				c:=I;
				exit when chiavi(I);
			end loop;
			libere:=libere-1;
			chiavi(c):=False;
			return c;
		end GetChiave;

	begin
		-- inizializzazione tutti armadietti liberi
		for I in ChiaveT loop
			chiavi(I):=True;
		end loop;

		loop
			select
				when libere>0 => accept acquistaBiglietto(tariffa: in TipoTariffa; chiave: out ChiaveT) do
					chiave:=GetChiave;
					Put_Line("Biglietteria: venduto biglietto (" & TipoTariffa'Image(tariffa) & "), chiave " & ChiaveT'Image(chiave));
				end;
			or 	accept restituisciChiave(chiave: in ChiaveT) do
					libere:=libere+1;
					chiavi(chiave):=True;
					Put_Line("Biglietteria: restituita chiave " & ChiaveT'Image(chiave));
				end;
			end select;
		end loop;
	end Biglietteria;

	-- PISCINA
	task Piscina is
		entry entra(CatUtente)(ID: in UtenteT);
		entry entraAbb(CatUtente)(ID: in UtenteT); -- prioritari all'interno di ogni categoria
		entry esce(ID: in UtenteT; tipo: in CatUtente);
	end;
	task body Piscina is
		cont : array(CatUtente) of Integer; -- un contatore per ogni tipo (Stud, Ord)

 		function getTotUtenti return Integer is
		begin
			return cont(Studente)+cont(Ordinario);
		end getTotUtenti;

		function categoriaMin return CatUtente is
		begin
			if cont(Studente)<=cont(Ordinario) -- in caso di parità privilegiati gli studenti
				then return Studente;
				else return Ordinario;
			end if;
		end categoriaMin;

		procedure PrintStatus is
		begin
			Put_Line("S=" & Integer'Image(cont(Studente)) & ", O=" & Integer'Image(cont(Ordinario)) & " -> " & CatUtente'Image(categoriaMin));
		end PrintStatus;

	begin
		-- inizializzazione contatori
		for I in CatUtente loop
			cont(I):=0;
		end loop;

		loop
			select
				when getTotUtenti<MAX and (Studente=categoriaMin or entraAbb(Ordinario)'Count+entra(Ordinario)'Count=0) => accept entraAbb(Studente)(ID: in UtenteT) do
					cont(Studente):=cont(Studente)+1;
					Put_Line("Piscina: entrato " & UtenteT'Image(ID) & " studente abbonato");
					PrintStatus;
				end;
			or	when getTotUtenti<MAX and (
                        (Studente=categoriaMin and entraAbb(Studente)'Count=0) or
                        entraAbb(Studente)'Count+entraAbb(Ordinario)'Count+entra(Ordinario)'Count=0
                    ) => accept entra(Studente)(ID: in UtenteT) do
					cont(Studente):=cont(Studente)+1;
					Put_Line("Piscina: entrato " & UtenteT'Image(ID) & " studente NON abbonato");
					PrintStatus;
				end;
			-- casi speculari per utente ordinario
			or	when getTotUtenti<MAX and (Ordinario=categoriaMin or entraAbb(Studente)'Count+entra(Studente)'Count=0) => accept entraAbb(Ordinario)(ID: in UtenteT) do
					cont(Ordinario):=cont(Ordinario)+1;
					Put_Line("Piscina: entrato " & UtenteT'Image(ID) & " ordinario abbonato");
					PrintStatus;
				end;
			or	when getTotUtenti<MAX and (
                        (Ordinario=categoriaMin and entraAbb(Ordinario)'Count=0) or
                        entraAbb(Ordinario)'Count+entraAbb(Studente)'Count+entra(Studente)'Count=0
                    ) => accept entra(Ordinario)(ID: in UtenteT) do
					cont(Ordinario):=cont(Ordinario)+1;
					Put_Line("Piscina: entrato " & UtenteT'Image(ID) & " ordinario NON abbonato");
					PrintStatus;
				end;
			or  	accept esce(ID: in UtenteT; tipo: in CatUtente) do
	 				cont(tipo):=cont(tipo)-1;
					Put_Line("Piscina: uscito " & CatUtente'Image(tipo) & " " & UtenteT'Image(ID));
					PrintStatus;
				end;
			end select;
		end loop;
	end Piscina;

	-- UTENTI
	task type Utente(ID: UtenteT; tipo: CatUtente; tariffa: TipoTariffa);
	task body Utente is
		chiave : ChiaveT;
	begin
		Biglietteria.acquistaBiglietto(tariffa, chiave);
		Put_Line(CatUtente'Image(tipo) & " (" & TipoTariffa'Image(tariffa) & ") in attesa di entrare...");

		if tariffa=Abbonato
			then Piscina.entraAbb(tipo)(ID);
			else Piscina.entra(tipo)(ID);
		end if;
		delay 2.0;

		Piscina.esce(ID, tipo);
		Biglietteria.restituisciChiave(chiave);

	end Utente;

	type UtenteAcc is access Utente;
	U : UtenteAcc;
begin

	-- main: creazione/attivazione utenti
	for I in UtenteT loop
		case I mod 4 is
			when 0 => U := new Utente(I, Studente, Abbonato);
			when 1 => U := new Utente(I, Studente, NonAbbonato);
			when 2 => U := new Utente(I, Ordinario, Abbonato);
        		when others => U := new Utente(I, Ordinario, NonAbbonato);
		end case;
	end loop;

end PiscinaPubblica;
