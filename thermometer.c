#define TERMOMETER 0x86
#define FURNACE 0x87
#define ENGAGE 1
#define DISENGAGE 0
void Regulate(double minTemp, double maxTemp)
{
    for (;;)
    {
        while (in(THERMOMETER) > minTemp)
            wait(1);
        out(FURNACE, ENGAGE);

        while (in(THERMOMETER) < maxTemp)
            wait(1);
        out(FURNACE, DISENGAGE);
    }
}