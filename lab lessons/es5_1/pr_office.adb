with Ada.Text_IO;
with Ada.Integer_Text_IO;
with Ada.Numerics.Float_Random;

use Ada.Text_IO;
use Ada.Integer_Text_IO;

procedure Pr_Office is

    -----------------------------------------------------------------------
    -- TYPES
    -----------------------------------------------------------------------
    type Service_Type is (TUR, EVE);
    type Client_ID_Type is range 1 .. 20;

    Num_Counters : constant Integer := 5;

    -----------------------------------------------------------------------
    -- SERVER TASK
    -----------------------------------------------------------------------
    -- Manages the counters. TUR clients have priority.
    task Office_Manager is
        entry Acquire_TUR (Client_ID : in Client_ID_Type);
        entry Acquire_EVE (Client_ID : in Client_ID_Type);
        entry Release_Counter (Client_ID : in Client_ID_Type);
    end Office_Manager;

    -----------------------------------------------------------------------
    -- CLIENT TASK TYPE
    -----------------------------------------------------------------------
    -- Note: A task type is NOT a tagged type!
    -- Therefore the correct access type is: access Client
    task type Client (ID : Client_ID_Type; Requested_Service : Service_Type);
    type Client_Access is access Client;

    -----------------------------------------------------------------------
    -- SERVER TASK BODY
    -----------------------------------------------------------------------
    task body Office_Manager is
        Occupied_Counters : Integer := 0;
    begin
        Put_Line("OFFICE: Opened with" & Integer'Image(Num_Counters) & " counters available.");

        loop
            select
                -- Priority: TUR (urgent service)
                when Occupied_Counters < Num_Counters and then Acquire_TUR'Count > 0 =>
                    accept Acquire_TUR (Client_ID : in Client_ID_Type) do
                        Put_Line ("URP [TUR]: Client" & Client_ID_Type'Image(Client_ID) & " acquires a counter.");
                        Occupied_Counters := Occupied_Counters + 1;
                    end Acquire_TUR;

            or
                -- If no TUR pending, accept EVE clients
                when Occupied_Counters < Num_Counters =>
                    accept Acquire_EVE (Client_ID : in Client_ID_Type) do
                        Put_Line ("URP [EVE]: Client" & Client_ID_Type'Image(Client_ID) & " acquires a counter.");
                        Occupied_Counters := Occupied_Counters + 1;
                    end Acquire_EVE;

            or
                -- Release always accepted
                accept Release_Counter (Client_ID : in Client_ID_Type) do
                    Put_Line("URP: Client" & Client_ID_Type'Image(Client_ID) & " releases a counter.");
                    Occupied_Counters := Occupied_Counters - 1;
                end Release_Counter;

            end select;
        end loop;
    end Office_Manager;

    -----------------------------------------------------------------------
    -- CLIENT TASK BODY
    -----------------------------------------------------------------------
    task body Client is
        Gen : Ada.Numerics.Float_Random.Generator;
        Stay_Duration : Duration;
    begin
        Ada.Numerics.Float_Random.Reset(Gen);

        -- Random time between 0.5 and 2.5 seconds
        Stay_Duration := Duration(3) * 2.0 + 0.5;

        Put_Line ("Client" & Client_ID_Type'Image(ID) & ": Requesting " & Service_Type'Image(Requested_Service));

        if Requested_Service = TUR then
            Office_Manager.Acquire_TUR(ID);
        else
            Office_Manager.Acquire_EVE(ID);
        end if;

        Put_Line ("Client" & Client_ID_Type'Image(ID) & ": At the counter for" & Duration'Image(Stay_Duration));
        delay Stay_Duration;

        Office_Manager.Release_Counter(ID);

        Put_Line ("Client" & Client_ID_Type'Image(ID) & ": Finished and leaving.");
    end Client;

begin
    Put_Line("--------------------------------------------------");
    Put_Line("Public Relations Office Simulation (Ada)");
    Put_Line("Priority: TUR > EVE");
    Put_Line("--------------------------------------------------");

    ---------------------------------------------------------------------
    -- CLIENT CREATION
    ---------------------------------------------------------------------
    declare
        C1, C2, C3 : Client_Access;
    begin
        C1 := new Client (1, EVE);
        delay 0.1;
        C2 := new Client (2, EVE);
        delay 0.1;
        C3 := new Client (3, TUR);
    end;

end Pr_Office;
