from data_loader import cargar_datos
from modelo_bayes import entrenar_modelo
from prediccion import evaluar_cliente

# Cargar dataset
datos = cargar_datos()

# Entrenar modelo
modelo = entrenar_modelo(datos)

# Evaluar nuevo cliente
evaluar_cliente(modelo)