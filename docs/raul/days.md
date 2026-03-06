Día 1 — eBPF Observer funcionando
El objetivo es simple: al final del día ves en tu terminal qué archivos abre cada proceso en tiempo real.
Setup del entorno

Instalar dependencias del sistema (clang, llvm, libelf-dev, libbpf-dev, linux-headers, build-essential)
Verificar que BTF existe: ls /sys/kernel/btf/vmlinux
Verificar Go 1.22+: go version
Instalar ollama y bajar el modelo: ollama pull qwen2.5:7b
Crear la estructura del proyecto e inicializar el módulo Go

El programa eBPF

Escribir ebpf/observer.bpf.c con tracepoint en sys_enter_openat
Compilar con clang a bytecode: clang -O2 -g -target bpf ...
Verificar que el .bpf.o se generó sin errores

El daemon en Go

Agregar github.com/cilium/ebpf al módulo
Escribir observer/types.go con el struct Event (layout idéntico al C)
Escribir observer/daemon.go que carga el objeto eBPF, attacha el tracepoint, y lee el ring buffer
Escribir main.go mínimo que arranca el daemon e imprime eventos en stdout

Verificación

sudo ./godshell corre sin errores
Abrís otro terminal, corrés ls, y ves el evento aparecer en godshell
Abrís Chrome o cualquier app y ves el stream de archivos que abre
