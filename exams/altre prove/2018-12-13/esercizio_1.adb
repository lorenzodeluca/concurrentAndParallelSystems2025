--------------------------------------------------------------------------------
--  *  Prog name esercizio1.adb
--  *  Project name esercizio1
--  *
--  *  Version 1.0
--  *  Last update 18/12/13
--  *
--  *  Created by anna ciampolini on 18/12/13.
--  *  Copyright (c) 2013 __MyCompanyName__.
--  *  All rights reserved.
--  *    or (keep only one line or write your own)
--  *  GNAT modified GNU General Public License
--  *
--------------------------------------------------------------------------------

with Ada.Text_IO, Ada.Integer_Text_IO;
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure esercizio1 is
 
   type cliente_ID is range 1..30;		-- numero veicoli

	type corsia is (N, S);
	type direzione is (NORD, SUD);
	
task type cliente (ID: cliente_ID; D:direzione);

type ac is access cliente;

task type server is
 entry entra_daN (ID: in cliente_ID;  C: out corsia);  
 entry entra_daS (ID: in  cliente_ID);   
 entry esceN(ID: in  cliente_ID; D: in direzione); 
 entry esceS(ID: in cliente_ID); 
end server;
 
SERV: server;  -- creazione task server

 

  ------ definizione PROCESSO server: 
   task body server   is
   MAX  : constant INTEGER := 5; -- capacita' corsie
       
   incorsia: array(corsia'Range) of Integer;
   inCN:array(direzione'Range) of Integer;
   
   
   begin
   Put_Line ("SERVER iniziato!");
   --INIZIALIZZAZIONI:
  	for i in corsia'Range loop          
      incorsia(i):=0;
     end loop;
	 
	 for i in direzione'Range loop          
      inCN(i):=0;
     end loop;

    --Gestione richieste
	delay 2.0;
	loop
		select
			when  incorsia(S)<MAX   =>  --ingresso da nord su corsia sud 
				accept entra_daN(ID: in cliente_ID; C: out corsia) do
					Put_Line("entra in corsia sud il veicolo "& cliente_ID'Image(ID) &" !");
					incorsia(S):=incorsia(S)+1;
					C:=S;
				end;         
			or 
			when  incorsia(S)=MAX and incorsia(N)<MAX and inCN(SUD)=0 and entra_daS'COUNT=0   =>  -- ingresso da nord su corsia  nord
				accept entra_daN(ID: in cliente_ID; C: out corsia) do
					Put_Line("entra in corsia nord il veicolo da NORD "& cliente_ID'Image(ID) &" !");
					incorsia(N):=incorsia(N)+1;
					inCN(NORD):=inCN(NORD)+1;
					C:=N;
				end;            
			or
				when  incorsia(N)<MAX and inCN(NORD)=0   =>  -- ingresso da sud su corsia nord 
				accept entra_daS(ID: in cliente_ID) do
					Put_Line("entra in corsia nord il veicolo da S "& cliente_ID'Image(ID) &" !");
					incorsia(N):=incorsia(N)+1;
					inCN(SUD):=inCN(SUD)+1;
				end;         
			or
				accept esceS(ID: in cliente_ID) do
					Put_Line("esce dalla corsia sud il veicolo"& cliente_ID'Image(ID) &" !");
					incorsia(S):=incorsia(S)-1;
				end;           
			or
				accept esceN(ID: in cliente_ID;D: in direzione) do
					Put_Line("esce dalla corsia nord il veicolo"& cliente_ID'Image(ID) &" !");
					incorsia(N):=incorsia(N)-1;
					inCN(D):=inCN(D)-1;
				end;       
			end select;
	end loop;
end;
  
 ------------------processi clienti:
 
task body cliente is
C:corsia;   
begin
   
   Put_Line ("veicolo" & cliente_ID'Image (ID) & " in direzione"&  direzione'Image(D) &" iniziato!");
   if D=NORD
   then 	SERV. entra_daN(ID,C);
			delay 2.0;
			if C=N
			then SERV.esceN(ID,NORD);
			else SERV.esceS(ID);
			end if;
	else
			SERV. entra_daS(ID);
			delay 2.0;
			SERV.esceN(ID, SUD);
			end if;

end;

------------------------------- main: 
   NewV: ac;

begin -- equivale al main
		
   for I in cliente_ID'Range 
   loop  -- ciclo creazione task OP
      NewV := new cliente (I, NORD); 
	  NewV := new cliente (I, SUD); 
    end loop;


end esercizio1;
