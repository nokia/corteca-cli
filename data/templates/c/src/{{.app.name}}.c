#include <stdio.h>
#include <time.h>
#include <unistd.h>

/*
    {{ .app.title }}
    author: {{ .app.author }}
*/

//{{if .app.options.include_libhlapi}}#include <libhlapi.h>{{end}}

int main() {
  printf("Hello from %s\n", "{{.app.title}}");

  for (;;) {
    sleep(5);
  }

  return 0;
}
