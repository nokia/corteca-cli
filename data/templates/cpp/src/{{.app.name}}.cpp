#include <iostream>
#include <chrono>
#include <thread>

/*
    {{ .app.title }}
    author: {{ .app.author }}
*/

//{{if .app.options.include_libhlapi}}#include <libhlapi.h>{{end}}

int main()
{
    std::cout << "Hello World from " << "{{.app.title}}" << std::endl;
    for (;;) {
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }
    return 0;
}
