# TP0: Docker + Comunicaciones + Concurrencia

Solución al TP0 de la materia Sistemas Distribuidos. Cada ejercicio esta resuelto en su propia rama. Debajo de cada enunciado se encontrará información pertinente a la resolución de cada uno, en caso de que sea necesario.

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
