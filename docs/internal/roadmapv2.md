# GODSHELL — ROADMAP DÍAS 3-7
> Sin código. Solo decisiones técnicas, foco, y criterios de éxito.

---

## ESTADO ACTUAL (Días 1-2 ✅)
eBPF daemon con 4 tracepoints, context engine, FrozenSnapshot, LLM bridge con 14 tools, TUI con Tool Cards, SQLite persistence, y el demo funcional de fileless malware C2 detection via heap. La base está sólida.

---

## DÍA 3 — Browser Map

### Browser Map
Pure Go, cero eBPF nuevo. Parseas `/proc/*/cmdline` buscando procesos con flags de Chromium o Firefox y extraes el árbol completo.
Para Chrome: qué PID es browser, cuál es renderer, cuál es GPU, y qué URL corre en cada renderer via `--site-for-process`.
Para Firefox: el flag de URL no existe así que dependemos de identificar los roles (`-contentproc`, Web Content, Socket Process). El Socket Process es clave porque por ahí pasa todo el SSL.
Es la única tool app-specific de todo el sprint. El resto es agnóstico.

**Criterio de done:** El árbol de navegadores muestra URLs por tab (Chrome) o roles de red (Firefox).

---

## DÍA 4 — SSL Interception

### Decisión de scope — Chrome fuera de v1
Chrome bundlea BoringSSL compilado estáticamente y el binario está stripped. Encontrar los offsets de SSL_write/SSL_read requiere trabajo de reversing por versión de Chrome. Es un rabbit hole que no cabe en el sprint. Se anuncia como v2 en el README con una línea técnica honesta. Los reviewers de HackerNews van a entender exactamente por qué.

### Lo que sí cubres — y es más que suficiente

**OpenSSL vía libssl.so:** Cubre la mayoría absoluta del ecosistema: curl, wget, Python, Ruby, PHP, Rust con openssl crate, Node.js en la mayoría de configs, y prácticamente cualquier CLI tool del sistema. Esta librería tiene símbolos, el attach es directo.

**Firefox vía libnss3.so:** Firefox usa NSS en vez de OpenSSL, pero también tiene símbolos. Las funciones objetivo son `PR_Write` y `PR_Read` en vez de los equivalentes de OpenSSL, pero el patrón de implementación es idéntico. Firefox cubierto en 2 horas adicionales sobre lo que ya tienes.

**Go binaries vía crypto/tls:** Los binarios Go no usan libssl — compilan crypto/tls directamente. Pero casi todos los binarios Go tienen símbolos porque Go los incluye por default. Puedes detectar binarios Go automáticamente buscando el Go build ID en el ELF y attachar a los símbolos de crypto/tls. Esto cubre todo tu stack de microservicios y cualquier herramienta CLI escrita en Go.

### HTTP Reconstructor
El mayor riesgo técnico del día. Los bytes SSL que capturas no llegan en frames HTTP completos — el kernel puede partir una request en múltiples eventos. Si pasas esos bytes directamente a un parser HTTP vas a tener drops. La solución es un buffer acumulativo por PID que intenta parsear, y si falla por datos incompletos simplemente espera más eventos antes de intentar de nuevo. HTTP/1.1 primero, que cubre la mayoría del tráfico de desarrollo. HTTP/2 es deseable pero no es bloqueante para el demo.

La parte más valiosa del reconstructor no es parsear headers — es lo que haces después: normalizar paths dinámicos a templates, detectar auth headers automáticamente, identificar endpoints sin autenticación, y construir el APIMap que el LLM puede describir en lenguaje natural.

**Criterio de done:** El LLM puede decir "este proceso hizo estas requests a esta API con este auth token" para cualquier proceso que use libssl, libnss3, o Go crypto/tls. Si eso funciona con curl y Firefox, el día terminó.

---

## DÍA 5 — Human Navigation Layer

### El concepto
El humano tiene intuición. El agente tiene acceso al kernel. Este feature conecta los dos. El usuario navega el árbol de procesos con el teclado, selecciona un proceso, y ese proceso se convierte en contexto implícito para el LLM sin que el usuario tenga que mencionarlo. La interacción parece telepática.

### Layout
Panel izquierdo con el árbol de procesos navegable. Panel derecho con el chat. El proceso seleccionado se inyecta silenciosamente en el contexto del LLM. Cuando presionas un hotkey sobre un proceso, la Tool Card aparece en el chat antes de que hayas escrito nada.

### Los hotkeys que importan
Navegar con j/k, seleccionar con Enter, y shortcuts directos para las tools más comunes: ssl_intercept, get_maps, shell_history. El criterio de diseño es que la investigación más común — "qué está haciendo este proceso" — no requiera escribir nada.

### Por qué este feature importa para el demo
El video del día 7 muestra a alguien navegando el árbol, presionando una tecla sobre el renderer de banking.com, y viendo aparecer las Tool Cards automáticamente. Ninguna herramienta de seguridad existente se ve así. Es el momento cinematográfico del demo.

**Criterio de done:** Seleccionar un proceso inyecta contexto. Los hotkeys disparan tools. El LLM responde con conocimiento del proceso seleccionado sin que el usuario lo mencione explícitamente.

---

## DÍA 6 — Threat Intel + Trace Rewrite + README

### Syscall Trace (Opcional si hay tiempo)
Reescribir el wrapper de strace con eBPF real. El objetivo no es feature parity total — es deshacerte de la dependencia externa. Strace decodifica los pointers a strings leyendo memoria; tu eBPF captura syscall numbers limpiamente. Para los clave (`openat`, `execve`, `connect`) resuelves los strings leyendo memoria desde el probe. Para el resto, el nombre del syscall y el error code es suficiente narrativa para el LLM.

### Threat Intel
Dos APIs, nada más: VirusTotal para hashes y dominios, AbuseIPDB para IPs. Lo importante no es la implementación — son llamadas REST simples — sino dónde se integra. La integración correcta es automática y silenciosa: cuando el snapshot se genera, las IPs de conexiones salientes ya vienen flaggeadas si tienen reportes. Cuando el LLM hashea un binario, el resultado ya incluye el veredicto de VT sin que el LLM tenga que pedirlo. El usuario nunca ve la integración, solo ve que Godshell ya sabe que esa IP tiene 847 reportes de abuse.

El rate limit de VT en free tier es 4 requests por minuto. La mitigación es cachear agresivamente y solo hacer lookups cuando el usuario interactúa con un proceso específico, no en background para todos los procesos del snapshot.

### README
Este es el entregable más importante del día. Más importante que cualquier feature. El README actual describe capacidades. El nuevo empieza con el momento: el proceso que no estaba en disco, el heap dump, el C2 config en plaintext. Los GIFs van antes del primer párrafo de descripción. La sección de install son máximo 3 comandos. Chrome aparece en el Roadmap como v2 con una línea técnica, no como disculpa.

La regla de oro: alguien que llega al README sin contexto previo debería entender en 10 segundos qué hace Godshell y por qué es diferente. Si tienen que leer bullets para entenderlo, el README falló.

**Criterio de done:** El README abre con los GIFs. La historia del fileless malware está en las primeras 10 líneas. Install en 3 comandos.

---

## DÍA 7 — Demo + Lanzamiento

### Cero código
Este día es completamente de producción del demo y ejecución del lanzamiento. Si hay bugs que arreglar, se arreglan antes de este día.

### Demo 1 — Fileless Malware (60 segundos)
Ya tienes esto funcionando. Terminal limpia, font grande, el proceso masqueradeando como zsh, el heap dump, el C2 config. Sin narración. El output del LLM habla solo. Termina con tres líneas de texto en pantalla.

### Demo 2 — SSL Interception / API Map (90 segundos)
Firefox abierto con cualquier webapp. Godshell en split screen. chrome_map muestra el árbol. Navegas al renderer con el teclado, presionas el hotkey de ssl_intercept. Usas la app normalmente por 10 segundos. El API Map aparece: endpoints, auth tokens, flags de seguridad. El mensaje visual es que el browser no sabe que lo estás mirando.

Opcionalmente, si tienes un backend Node o Python corriendo localmente, mostrar el stack completo — frontend en Firefox y backend simultáneamente — es más impresionante que solo el browser.

### Timing del post en HackerNews
Martes o miércoles, 9-10am hora del este. Es el slot con mayor tráfico y menor competencia. El título no hace hype: "Show HN: Godshell – eBPF + LLM agent that reads your kernel to investigate processes and intercept TLS". Los GIFs hacen el trabajo, no el título.

Los primeros 30 minutos determinan si rankea. Tener 3-4 personas listas para comentar técnicamente desde el lanzamiento — preguntas reales sobre implementación — genera el engagement que el algoritmo de HN necesita para darle visibilidad.

**Twitter/X simultáneo** con el Demo 2 embedido, taggeando cuentas de infosec y eBPF. r/netsec y r/linux como backup si HN no rankea el primer día.

---

## RESUMEN

| Día | Foco | Lo que no es negociable |
|-----|------|------------------------|
| 3 | Browser map puro | Árbol navegable con URLs de Chrome |
| 4 | SSL interception libssl + NSS + GoTLS | Demo funciona con Firefox y curl end-to-end |
| 5 | Human navigation TUI | Hotkeys disparan tools sin escribir |
| 6 | Threat intel + README + Trace Rewrite | GIFs antes del fold, install en 3 comandos |
| 7 | Grabar + lanzar | Cero código nuevo |

---

## DECISIONES TÉCNICAS FINALES

**Chrome:** v2 con offset database. No es una limitación, es scope correcto.

**HTTP/2:** Deseable pero no bloqueante. Si HTTP/1.1 parsea bien y el 70% del tráfico funciona, el demo es válido.

**YARA, containers, timeline forense:** Roadmap. Se mencionan en el README como próximos features, no se implementan esta semana.

**strace wrapper:** Se reemplaza con eBPF real pero sin pretender feature parity total. El LLM necesita narrativa, no exhaustividad.