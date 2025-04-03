#include <stdio.h>
#include <unistd.h>
#include <time.h>

/*
    {{ .app.name }}
    author: {{ .app.author }}
*/


int main()
{

   printf("Hello from %s\n", "{{.app.name}}");

   for (;;)
   {
      sleep(5);
   }

   return 0;
}
