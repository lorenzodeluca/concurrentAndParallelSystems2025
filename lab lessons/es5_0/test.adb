with Ada.Text_IO, Ada.Integer_Text_IO;  -- Importazione dei pacchetti
use Ada.Text_IO, Ada.Integer_Text_IO;

procedure Task1 is  -- procedura principale
    task First_Task;  -- dichiarazione del primo processo

    task body First_Task is
    begin  -- corpo della task
        for Index in 1..4 loop
            Put("This is in First_Task, pass number ");
            Put(Index, 3);
            New_Line;
        end loop;
    end First_Task;

    task Second_Task;  -- dichiarazione del secondo processo

    task body Second_Task is
    begin  -- corpo della task
        for Index in 1..7 loop
            Put("This is in Second_Task, pass number ");
            Put(Index, 3);
            New_Line;
        end loop;
    end Second_Task;

begin  -- corpo della procedura main
    Put_Line("Questo è il main task..");
end Task1;
