with Ada.Text_IO;
use Ada.Text_IO;

procedure HotDog is
    task Gourmet is  -- dichiarazione della task
        entry Make_A_Hot_Dog;
    end Gourmet;

    task body Gourmet is  -- definizione del corpo della task
    begin
        for Index in 1..4 loop
            accept Make_A_Hot_Dog do  -- accetta la richiesta di fare un hot dog
                delay 0.8;  -- simula una pausa, come il "sleep"
                Put("Metto hot dog nel pane..");
                New_Line;
                Put_Line("Aggiungo senape");
            end Make_A_Hot_Dog;
        end loop;
    end Gourmet;
begin
    for Index in 1..4 loop
        Gourmet.Make_A_Hot_Dog; --entry call
        delay 0.1;
        Put_Line("Mangio l’hot dog");
        New_Line;
    end loop;
end HotDog;
