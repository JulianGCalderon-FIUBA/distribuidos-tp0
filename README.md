# TP0: Docker + Comunicaciones + Concurrencia

Solución al TP0 de la materia Sistemas Distribuidos. Cada ejercicio esta resuelto en su propia rama. Debajo de cada enunciado se encontrará información pertinente a la resolución de cada uno, en caso de que sea necesario.

```bash
> git ls-remote --refs -q | head -n -1 | awk -F"/|\t" '{printf "- %s: %s\n", $4, $1}'
- ej1: dac8546cc3e418c5fac3583db57b4f90fb10d6df
- ej2: 6b9f5c2c46bf02ac61074d9dfbe4b203fe929a79
- ej3: 4ef16536aa3113e89566ee2d2a760f73359d6c8e
- ej4: b5f968b44e189b936e67bb3f8892bc2226a7ea6f
- ej5: 843cb85ccbdfa6199ba313633038a3878ff25bad
- ej6: 06c48eee729531fccb55dfb2cd6a7a21e61447a4
- ej7: 4e612af3333c9e309d9eb8e522c7c7cc84ed128d
- ej8: 0dd030bc9dddb8a0d230e524bd4df3b944145912
```

## Ejercicio N°1:

>  Definir un script de bash `generar-compose.sh` que permita crear una definición de DockerCompose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

Para resolver este ejercicio, se definio un script de shell `generar-compose.sh` que utiliza el builtin `echo` para escribir el archivo DockerCompose deseado.

## Ejercicio N°2:

> Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera un nuevo build de las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida afuera de la imagen (hint: `docker volumes`).

Para evitar tener que buildear las imagenes de Docker al cambiar los archivos de configuracion, entonces:
1. Ignore los repectivos archivos a partir de un archivo `.dockerignore`
2. Configure un bind mount desde Docker Compose para montar los archivos de configuracion existentes en la ubicacion esperada dentro de los contenedores.

## Ejercicio N°3:

> Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un EchoServer, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado. En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir: `action: test_echo_server | result: fail`. El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina host y no se puede exponer puertos del servidor para realizar la comunicación (hint: `docker network`).

Para lograr verificar el correcto funcionamiento del servidor sin tener `nc` instalado ni exponer puertos, es necesario ejecutar la verificacion desde un contenedor de docker dentro de la misma `network`.

Luego, la verificacion corre un contenedor de docker con la imagen `alpine` y ejecuta un shell script. Este script utiliza `nc` y `diff` para comparar la respuesta del servidor con el resultado esperado.

## Ejercicio N°4:

> Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

Antes de comenzar a resolver el ejercicio, migré todo el servidor a Go. Esto sirvio como entrada en calor para el lenguaje.

Para la implementacion del graceful shutdown, utlice el patron `context` de Go. En determinadas partes del codigo, se verifica que el contexto no haya finalizado. Si finaliza, entonces libera los recursos y retorna.
- Desde el lado del cliente, se verifica el contexto despues de cada conexion 
- Desde el lado del servidor, se utiliza una version no bloqueante del `accept`, y se verifica el contexto cada 0.5s. Esto permite que el servidor no se quede esperando a un cliente que nunca va a llegar.

Para asegurarme de que siempre se liberan los recursos, utilice la instruccion de go `defer`. Esta asegura que se ejecuten los destructores incluso ante un `panic`.

## Ejercicio N°5:

> Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso:
>
> El cliente emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente. Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.
>
> El servidor emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
> Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.
>
> Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
> * Definición de un protocolo para el envío de los mensajes.
> * Serialización de los datos.
> * Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
> * Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


Antes de comenzar el ejercicio, migre el archivo de `utils` de Python a Go para poder usarlo. Mantuve la misma interfaz para no modificar el ejercicio. El mismo se encuentra en el modulo [server/lottery](server/lottery/lottery.go).

Los datos de la apuesta se pueden especificar desde el archivo de configuracion:
```yaml
bet:
  firstName: "laura"
  lastName: "lopez"
  document: 40000001
  birthdate: "2001-05-01"
  number: 1
```
Tambien se pueden especificar con variables de entorno:
```bash
export NOMBRE="laura"
export APELLIDO="laura"
export DOCUMENTO=40000001
export NACIMIENTO="2001-05-01"
export NUMERO=1
```

La mensajes intercambiados entre cliente-servidor tienen formato CSV. Este es simple y permite separar correctamente mensajes con saltos de linea (\n). Ademas, permite reutilizar codigo ya que ya era utilizado para la serializacion de apuestas en disco.

El protocolo sigue la siguiente secuencia:
1. **Client**: Se conecta al servidor
1. **Client**: Envia un mensaje de HELLO con su ID
1. **Cliente**: Envia su apuesta
1. **Servidor**: Guarda la apuesta en disco
1. **Servidor**: Envia un mensaje de confirmacion OK
1. **Servidor**: Espera al siguiente cliente

La logica de negocio se encuentra en:
- [common/bet](common/bet.go): Contiene el tipo `LocalBet` y como serializarlo a un registro de CSV
- [server/lottery](server/lottery/lottery.go): Contiene la logica de guardado de las apuestas a disco

La logica de comunicacion para cada entidad se encuentra, respectivamente, en:
- [server/server](server/server.go)
- [client/client](client/client.go)
- [common/protocol](common/protocol.go): Contiene los mensajes entre cliente-servidor, y como serializarlos

Para evitar el fenomeno conocido como `short read` es necesario continuar haciendo lecturas hasta encontrar un salto de linea. Afortunadamente, la biblioteca de serializacion de CSV ya realiza esto (utilizando internamente un buffered reader). Para evitar un `short write` entonces es necesario verificar que se hayan escrito todos los bytes necesarios. Podemos asegurarnos de esto utilizando el metodo `Flush` de la biblioteca de serializacion de CSV.

Para probar la correcta ejecucion del sistema, se puede ejecutar:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
make docker-compose-up
```

Luego, se puede observar el archivo CSV (aunque al tener el mismo archivo de configuracion, todos los registros son iguales))
```bash
> docker exec server cat bets.csv
5,laura,lopez,40000001,2001-05-01,1
1,laura,lopez,40000001,2001-05-01,1
4,laura,lopez,40000001,2001-05-01,1
3,laura,lopez,40000001,2001-05-01,1
2,laura,lopez,40000001,2001-05-01,1
```

## Ejercicio N°6:

> Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
> Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.
>
> En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.
>
> La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB.
>
> El servidor, por otro lado, deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

El protocolo sigue la siguiente secuencia:
1. **Client**: Se conecta al servidor
1. **Client**: Envia un mensaje de HELLO con su ID y la cantidad de apuestas a enviar
1. **Cliente**: Envia todas las apuestas del batch
1. **Servidor**: Guarda todas las apuestas a disco
1. **Servidor**: Envia un respuesta al cliente:
   - OK si se procesaron todas las apuestas correctamente
   - ERR si se encontro algun error
1. **Servidor**: Espera al siguiente cliente

Bajo este protocolo, cada apuesta ocupa como maximo 57bytes (longitud de la linea mas larga del dataset, contando el salto de linea). Luego, podemos usar un tamaño de batch de 140 = 8000 // 57, de modo que los paquetes no excedan los 8kb.

Para probar la correcta ejecucion del sistema, se puede ejecutar:
```bash
make docker-compose-up
```

Luego, se puede observar el archivo CSV
```bash
docker exec server cat bets.csv | less
```

Tambien podemos contar la cantidad de regitros guardados contando las lineas del archivos
```bash
> docker exec server wc -l bets.csv
78697 bets.csv
```

¿Y si queremos contar la cantidad de ganadores?
```bash
> docker exec server sh -c 'cat bets.csv | cut -d, -f6 | grep -o 7574 | wc -l'
10
```

## Ejercicio N°7:

> Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo. Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia. Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.
>
> El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`. Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo, no podrá responder consultas por la lista de ganadores. Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

Hasta ahora, el protocolo usa una conexion nueva por cada interaccion. Los clientes vuelven a conectarse si quieren seguir enviando información. Ahora, es necesario mantener los sockets activos durante toda la ejecucion. Esto requirió un refactor grande del lado del servidor.
- El servidor mantiene un array de conexiones, una por cada agencia.
    - Itera por cada conexion, resolviendo una peticion a la vez. Si no hay ninguna peticion, continua a la siguiente conexion.
    - Por cada ronda de peticiones, revisa si tiene una conexión entrante. Si no tiene ninguna conexion entrante, continua.

Debido a que la actualizacion al protocolo introduce nuevos mensajes. Tambien decidi invertir tiempo en refactorizar las estructuras usadas en la comunicacion en un nuevo paquete: [protocol](./protocol/protocol.go). Este paquete define las estructuras intercambiadas entre cliente-servidor, y como se serializan a `[]string`. El formato de los mensajes continua siendo un CSV, delimitados por saltos de linea. Un mensaje con forma `TIPO(Arg1, Arg2, ...)` se serializara como `TIPO,Arg1,Arg2,...\n`.

El protocolo sigue la siguiente secuencia:
1. **Client**: Se conecta al servidor
1. **Client**: Envia un mensaje de `HELLO(AgencyId)`
1. **Cliente**: Envia un mensaje de `BATCH(BatchSize)`
   1. **Cliente**: Envia todas las apuestas del lote, cada una a traves de un mensaje `BET(FirstName, LastName, Document, Birthdate, Number)`
   1. **Servidor**: Guarda todas las apuestas a disco
   1. **Servidor**: Envia un respuesta al cliente:
      - `OK()` si se procesaron todas las apuestas correctamente
      - `ERR()` si se encontro algun error
1. **Cliente**: Repite el paso 3 hasta haber enviado todas las apuestas
1. **Cliente**: Una vez envio todas las apuetas, envia un mensaje `FINISH()`

El servidor continua resolviendo peticiones concurrentemente hasta obtener un mensaje `FINISH` de cada cliente. Luego envia a cada agencia sus respectivos ganadores, a traves de un mensaje `WINNERS(Length, Document1, Document2, Document3, ...)`. Debido a que la cantidad de documentos es variable, incluimos la cantidad de ganadores.

Para probar el correcto funcionamiento del sistema, cree el siguiente script de valiacion. Este asegura que el archivo de apuestas almacenado en el servidor sea el agregado del archivo de apuestas de cada agencias. Se debe ejecutar una vez finalice el comando `make docker-compose-up`.
```
> ./validar-sistema.sh
OK
```

Si ejecutamos el sistema, podemos observar que los ultimos mensajes corresponden a los ganadores:
```bash
> make docker-compose-up
...
...
server   | 2024-09-01 00:05:47 INFO     action: sorteo | result: success
client1  | 2024-09-01 00:05:48 INFO     action: consulta_ganadores | result: success | cant_ganadores: 2
client3  | 2024-09-01 00:05:48 INFO     action: consulta_ganadores | result: success | cant_ganadores: 3
client2  | 2024-09-01 00:05:48 INFO     action: consulta_ganadores | result: success | cant_ganadores: 3
client5  | 2024-09-01 00:05:48 INFO     action: consulta_ganadores | result: success | cant_ganadores: 0
client4  | 2024-09-01 00:05:48 INFO     action: consulta_ganadores | result: success | cant_ganadores: 2
```

## Ejercicio N°8:

> Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

Hasta ahora, el servidor fue single-threaded. Se utiliza un único hilo de ejecucion y operaciones no bloqueantes para permitir resolver peticiones de los clientes de forma concurrente. En este ejercicio, finalmente se hace uso de las gorutinas para hacer que el servidor procese mensajes en paralelo.

El protocolo continua haciendo el mismo, lo unico que cambio fue el uso de las primitivas de sincronizacion:
- Es necesario utilizar un `Mutex` para asegurar que solo una gorutina escribe en el archivo de apuestas al mismo tiempo.
- Para asegurar que los ganadores se envien unicamente al final, se utiliza un `WaitGroup`.
- Para asegurar que no finalice la ejecucion hasta que todos los hilos hayan terminado, entonces se utiliza otro `WaitGroup`.
