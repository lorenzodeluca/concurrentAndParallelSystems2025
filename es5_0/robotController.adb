with Ada.Text_IO;               -- include text_io
with Ada.Integer_Text_IO;       -- include integer_text_io
use Ada.Text_IO;

procedure Task_Demo is

    -- Dichiarazione del tipo di task
    task type Intro_Task is
        entry Start;
        entry Turn_Left;
        entry Turn_Right;
        entry Stop;
    end Intro_Task;

    -- Corpo del task
    task body Intro_Task is
    begin
        accept Start;  -- attesa iniziale

        loop
            select
                accept Turn_Left;
                Put_Line("Turning left");

            or
                accept Turn_Right;
                Put_Line("Turning right");

            or
                accept Stop;
                Put_Line("Stop received");
                exit;  -- esce dal loop

            else
                Put_Line("Moving straight");
            end select;

            delay 0.5;
        end loop;
    end Intro_Task;

    -- Istanza del task “server”
    Task_1 : Intro_Task;

begin  -- task “client”
    Task_1.Start;
    delay 2.0;

    Task_1.Turn_Left;
    delay 2.0;

    Task_1.Turn_Right;
    delay 1.0;

    Task_1.Turn_Right;
    delay 2.0;

    Task_1.Stop;
end Task_Demo;
