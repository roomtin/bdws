#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
int main(int argc, char** argv)
{
  while(1) 
  {
    printf("Tick!\n");
    sleep(1);
  }
  return EXIT_SUCCESS;
}
