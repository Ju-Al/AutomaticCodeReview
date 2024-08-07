#include "divide.h"

#include <stdio.h>
#include <stdlib.h>


int main(int argc, char* argv[])
{
    long result;
    long test_case;

    if(argc < 1) {
        printf("Please specify testcase id.\n");
        return 1;
    } else {
        test_case = strtol(argv[1], NULL, 10);
        // Division by zero, only detected when ctu analysis is on
        result = divide(test_case, 0);
    }

    return (int)result;
}
