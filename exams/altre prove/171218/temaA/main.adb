with Ada.Numerics.Discrete_Random;
with Ada.Text_IO, Ada.Integer_Text_IO;
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure esame is
   type cliente_ID is range 1..15; -- 20 clienti
   type tipo is (under, over); -- tipo cliente

   --Implementazione codice random
   package tipo_random is new Ada.Numerics.Discrete_Random(tipo);
   use tipo_random;
   G : Generator;

   task type sportelli is
      entry ritiraEntra(tipo)(ID: in cliente_ID);
      entry ritiraEsce(tipo)(ID: in cliente_ID);
      entry restituisciEntra(tipo)(ID: in cliente_ID);
      entry restituisciEsce(tipo)(ID: in cliente_ID);
   end sportelli;

   S: sportelli;

   task type caveau is
      entry entra(tipo)(ID: in cliente_ID);
      entry esci(tipo)(ID: in cliente_ID);
   end caveau;

   C: caveau;

   task type cliente (ID: cliente_ID; eta: tipo);

   task body cliente is
   begin
      Put_Line ("cliente" & cliente_ID'Image (ID) & " (" & tipo'Image (eta) & ") iniziato");
      S. ritiraEntra(eta)(ID);
      delay 1.0;
      S. ritiraEsce(eta)(ID);
      C. entra(eta)(ID);
      delay 2.0;
      C. esci(eta)(ID);
      S. restituisciEntra(eta)(ID);
      delay 1.0;
      S. restituisciEsce(eta)(ID);
   end;

   task body sportelli
   is
   N:constant INTEGER := 5; --numero sportelli
   occupati: Integer;
   begin
      Put_Line ("sportelli iniziato!");
      occupati:=0;
      loop
         select
            when occupati < N =>
               accept ritiraEntra(over)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 chiede chiave!");
                  occupati:=occupati+1;
               end ritiraEntra;
         or
            accept ritiraEsce(over)(ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 ha ottenuto la chiave!");
               occupati:=occupati-1;
            end ritiraEsce;
         or
            when occupati < N and ritiraEntra(over)'COUNT=0 and restituisciEntra(over)'COUNT=0=>
               accept ritiraEntra(under)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 chiede chiave!");
                  occupati:=occupati+1;
               end ritiraEntra;
         or
            accept ritiraEsce(under)(ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 ha ottenuto la chiave!");
               occupati:=occupati-1;
            end ritiraEsce;
         or
            when occupati < N =>
               accept restituisciEntra(over)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 va a restituire la chiave!");
                  occupati:=occupati+1;
               end restituisciEntra;
         or
            accept restituisciEsce(over)(ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 se ne va!");
               occupati:=occupati-1;
            end restituisciEsce;
         or
            when occupati < N and ritiraEntra(over)'COUNT=0 and restituisciEntra(over)'COUNT=0=>
               accept restituisciEntra(under)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 va a restituire la chiave!");
                  occupati:=occupati+1;
               end restituisciEntra;
         or
            accept restituisciEsce(under)(ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 se ne va!");
               occupati:=occupati-1;
            end restituisciEsce;
         end select;
      end loop;
   end;

   task body caveau
   is
      occupato: Boolean;
      storicoOver: Integer;
      storicoUnder: Integer;
   begin
      Put_Line ("caveau iniziato!");
      occupato:=False;
      storicoOver:=0;
      storicoUnder:=0;
      loop
         select
            when occupato = False and (storicoOver<=storicoUnder or (storicoOver>storicoUnder and entra(under)'Count=0)) =>
             accept entra(over)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 ottiene cassetta dall'addetto ed entra!");
                  storicoOver:=storicoOver+1;
                  occupato:=True;
                  Put_Line ("UNDER70:" & Integer'Image (storicoUnder) & " OVER70: " & Integer'Image (storicoOver));
               end entra;
         or
            when occupato = False and (storicoUnder<storicoOver or (storicoUnder>=storicoOver and entra(over)'Count=0))=>
               accept entra(under)(ID : in cliente_ID) do
                  Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 ottiene cassetta dall'addetto ed entra!");
                  occupato:=True;
                  storicoUnder:=storicoUnder+1;
                  Put_Line ("UNDER70:" & Integer'Image (storicoUnder) & " OVER70: " & Integer'Image (storicoOver));
               end entra;
         or
            accept esci (over) (ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " over70 riconsegna cassetta all'addetto ed esce!");
               occupato:=False;
            end esci;
         or
            accept esci (under) (ID : in cliente_ID) do
               Put_Line ("cliente" & cliente_ID'Image (ID) & " under70 riconsegna cassetta all'addetto ed esce!");
               occupato:=False;
            end esci;
         end select;
      end loop;
   end;

   type ac is access cliente;
   New_cliente: ac;

begin -- equivale al main
   Reset(G);
   for I in cliente_ID'Range loop -- ciclo creazione task
      New_cliente := new cliente (I,Random(G));
   end loop;
end esame;
