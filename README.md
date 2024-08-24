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
