#include <stdio.h>
#include <time.h>
#include <unistd.h>

/*
    {{ .app.title }}
    author: {{ .app.author }}
*/

int main() {
  printf("Hello from %s\n", "{{.app.title}}");

  for (;;) {
    sleep(5);
  }

  return 0;
}
