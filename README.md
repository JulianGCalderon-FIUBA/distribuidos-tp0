# TP0: Docker + Comunicaciones + Concurrencia

Solución al TP0 de la materia Sistemas Distribuidos. Cada ejercicio esta resuelto en su propia rama. Debajo de cada enunciado se encontrará información pertinente a la resolución de cada uno, en caso de que sea necesario.

## Ejercicio N°1:

>  Definir un script de bash `generar-compose.sh` que permita crear una definición de DockerCompose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

Para resolver este ejercicio, se definio un script de shell `generar-compose.sh` que utiliza el builtin `echo` para escribir el archivo DockerCompose deseado.
