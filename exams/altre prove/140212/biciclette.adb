with Ada.Text_IO, Ada.Integer_Text_IO;
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure biciclette is
	type cliente_ID is range 1..10;
	type cliente is (esperto, nuovo);
	type bici is (donna, uomo, bambino);
	task type clienteT (ID: cliente_ID; Owned:bici);
	type acC is access clienteT;
	task type server is
		entry offri				(ID: cliente_ID; Off : bici); --deposita l'offerta per lo scambio, perde la bici, appena sarà presente una bici del tipo richiesto l'acquisirà, scambio in 2 tempi
		entry scambia_esperto	(bici)(ID: cliente_ID);
		entry scambia_nuovo		(bici)(ID: cliente_ID);
	end server;
	S: server;

	task body server is
		offerte: array(bici'Range) of Integer;
		begin
			Put_Line("Server Inizializzato");
			for i in bici'Range loop
				offerte(i):=0;
			end loop;
			delay 2.0;
			loop
				select
					accept offri(ID: cliente_ID; Off : bici) do --non vi sono vincoli per poter offrire una bici
						offerte(Off):=offerte(Off)+1;
						Put_Line("Offerta bici di tipo " & bici'Image(Off));
					end;
				or
					when offerte(donna)>0 =>		--vincolo per poter acquisire la bici
					accept scambia_esperto(donna)(ID: cliente_ID) do
						offerte(donna):=offerte(donna)-1;
						Put_Line("Utente esperto: Acquisita bici tipo Donna");
					end;
				or
					when offerte(uomo)>0 =>			--vincolo per poter acquisire la bici
					accept scambia_esperto(uomo)(ID: cliente_ID) do
						offerte(uomo):=offerte(uomo)-1;
						Put_Line("Utente esperto: Acquisita bici tipo Uomo");
					end;
				or
					when offerte(bambino)>0 =>		--vincolo per poter acquisire la bici
					accept scambia_esperto(bambino)(ID: cliente_ID) do
						offerte(bambino):=offerte(bambino)-1;
						Put_Line("Utente esperto: Acquisita bici tipo Bambino");
					end;
				or
					when offerte(donna)>0 and		--vincolo per poter acquisire la bici
					--vincoli di priorita' rispetto agli utenti esperti
					scambia_esperto(donna)'COUNT = 0 and
					scambia_esperto(uomo)'COUNT = 0 and
					scambia_esperto(bambino)'COUNT = 0 =>
					accept scambia_nuovo(donna)(ID: cliente_ID) do
						offerte(donna):=offerte(donna)-1;
						Put_Line("Utente nuovo: Acquisita bici tipo Donna");
					end;
				or
					when offerte(uomo)>0 and		--vincolo per poter acquisire la bici
					--vincoli di priorita' rispetto agli utenti esperti
					scambia_esperto(donna)'COUNT = 0 and
					scambia_esperto(uomo)'COUNT = 0 and
					scambia_esperto(bambino)'COUNT = 0 =>
					accept scambia_nuovo(uomo)(ID: cliente_ID) do
						offerte(uomo):=offerte(uomo)-1;
						Put_Line("Utente nuovo: Acquisita bici tipo Uomo");
					end;
				or
					when offerte(bambino)>0 and		--vincolo per poter acquisire la bici
					--vincoli di priorita' rispetto agli utenti esperti
					scambia_esperto(donna)'COUNT = 0 and
					scambia_esperto(uomo)'COUNT = 0 and
					scambia_esperto(bambino)'COUNT = 0 =>
					accept scambia_nuovo(bambino)(ID: cliente_ID) do
						offerte(bambino):=offerte(bambino)-1;
						Put_Line("Utente nuovo: Acquisita bici tipo Bambino");
					end;
				end select;
			end loop;
		end;
		task body clienteT is
		c : cliente;
		myBike : bici;
		req : bici;
		begin
			c := nuovo;
			myBike := Owned;
			loop				--clienti scambiano bici all'infinito
				delay 1.0;
				if myBike = donna then --designo la mia prossima bici
					req := uomo;
				elsif myBike = uomo then
					req := bambino;
				elsif myBike = bambino then
					req := donna;
				end if;
				S.offri(ID,myBike);	--offro la mia attuale bici
				if c=nuovo then
					S.scambia_nuovo(req)(ID);	--ottengo una nuova bici
				elsif c=esperto then
					S.scambia_esperto(req)(ID); --ottengo una nuova bici
				end if;
				myBike:=req;
				c := esperto;
			end loop;
		end;

		NewC: acC;
		begin
			for I in cliente_ID'Range
			loop
				NewC := new clienteT(I,donna);
				NewC := new clienteT(I,uomo);
				NewC := new clienteT(I,bambino);
			end loop;
end biciclette;
