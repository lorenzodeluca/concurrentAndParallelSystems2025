with Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Discrete_Random;
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure esame is

	type cliente_ID is range 1..30;

	type entrata is (entrata1, entrata2);
	type esito is (creditore, debitore);

task type cliente (ID: cliente_ID; ENT: entrata);

task type sala is
	entry entra (entrata) (ID: in cliente_ID);
	entry esce (ID: in  cliente_ID; E: out esito);
end sala;

task type cassa is
	entry salda (esito) (ID: in cliente_ID);
	entry lascia(ID: in cliente_ID);
end cassa;

S: sala;
C: cassa;



------ definizione processi server:
task body sala   is

	N  : constant INTEGER := 10;
	dentro : INTEGER := 0;
	entrati: array(entrata'Range) of Integer;
   	package Random_esito is new Ada.Numerics.Discrete_Random (esito);
   	use Random_esito;
	G : Generator;

	begin
	Put_Line ("SALA iniziato!");

	for i in entrata'Range loop
		entrati(i):=0;
	end loop;
	Reset(G);
	--Gestione richieste
	loop
		select
			when  dentro < N and entrati(entrata1) > entrati(entrata2) and entra(entrata2)'COUNT = 0  =>
				accept entra(entrata1)(ID: in cliente_ID) do
					Put_Line("Entra cliente "& cliente_ID'Image(ID) &" da entrata 1");
					entrati(entrata1):=entrati(entrata1)+1;
					dentro := dentro+1;
				end;
			or
			when  dentro < N and entrati(entrata1) <= entrati(entrata2) =>
				accept entra(entrata1)(ID: in cliente_ID) do
					Put_Line("Entra cliente "& cliente_ID'Image(ID) &" da entrata 1");
					entrati(entrata1):=entrati(entrata1)+1;
					dentro := dentro+1;
				end;
			or
			when  dentro < N and  entrati(entrata2) > entrati(entrata1) and entra(entrata1)'COUNT = 0  =>
				accept entra(entrata2)(ID: in cliente_ID) do
					Put_Line("Entra cliente "& cliente_ID'Image(ID) &" da entrata 2");
					entrati(entrata2):=entrati(entrata2)+1;
					dentro := dentro+1;
				end;
			or
			when  dentro < N and entrati(entrata2) <= entrati(entrata1) =>
				accept entra(entrata2)(ID: in cliente_ID) do
					Put_Line("Entra cliente "& cliente_ID'Image(ID) &" da entrata 2");
					entrati(entrata2):=entrati(entrata2)+1;
					dentro := dentro+1;
				end;
			or
				accept esce (ID: in  cliente_ID; E: out esito) do
					E := Random(G);
					Put_Line("Esce cliente "& cliente_ID'Image(ID) &" con esito: "& esito'Image(E));
					dentro := dentro-1;
				end;
			end select;
	end loop;
end;

task body cassa   is
	occupato  : INTEGER := 0;

	begin
	Put_Line ("CASSA iniziato!");
	--Gestione richieste
	loop
		select
			when  occupato = 0   =>
				accept salda (debitore) (ID: in cliente_ID) do
					Put_Line("Cliente "& cliente_ID'Image(ID) &" salda il suo debito");
					occupato := 1;
				end;
			or
			when  occupato = 0 and salda(debitore)'COUNT = 0 =>
				accept salda (creditore) (ID: in cliente_ID) do
					Put_Line("Cliente "& cliente_ID'Image(ID) &" riscuote il suo credito");
					occupato := 1;
				end;
			or
				accept lascia(ID: in cliente_ID) do
					Put_Line("Cliente "& cliente_ID'Image(ID) &" lascia il casinò");
					occupato := 0;
				end;
			end select;
	end loop;
end;

 ------------------processi clienti:

task body cliente is
E:esito;

begin

	S.entra(ENT)(ID);
	delay 3.0;
	S.esce(ID, E);
	delay 1.0;
	C.salda(E)(ID);
	delay 2.0;
	C.lascia(ID);

end;

------------------------------- main:
	type ac is access cliente;

   	CLIENT: ac;
	package Random_entrata is new Ada.Numerics.Discrete_Random (entrata);
	use Random_entrata;
	GE : Generator;

begin
	Reset(GE);
	for I in cliente_ID'Range
	loop
		CLIENT := new cliente (I, Random(GE));
	end loop;


end esame;



