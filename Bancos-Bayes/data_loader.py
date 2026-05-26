import pandas as pd

def cargar_datos():
    datos = pd.read_csv("dataset.csv", sep="\t")
    return datos

# Guardar datos en variable
datos = cargar_datos()

# Mostrar columnas
print(datos.columns)