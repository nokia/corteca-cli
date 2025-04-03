#include <iostream>
#include <chrono>
#include <thread>

/*
    {{ .app.name }}
    author: {{ .app.author }}
*/

int main()
{

    std::cout << "Hello World from " << "{{.app.name}}" << std::endl;
    for (;;) {
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }
    return 0;
}
