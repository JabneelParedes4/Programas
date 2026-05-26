import pandas as pd

def evaluar_cliente(modelo):

    ingreso = float(input("Ingreso mensual: "))
    deuda = float(input("Deuda actual: "))
    score = int(input("Score crediticio: "))
    edad = int(input("Edad: "))

    cliente = pd.DataFrame({
        "ingreso": [ingreso],
        "deuda": [deuda],
        "score": [score],
        "edad": [edad]
    })

    riesgo = modelo.predict(cliente)

    # Lógica simple préstamo
    if riesgo[0] == "Bajo":
        prestamo = ingreso * 4

    elif riesgo[0] == "Medio":
        prestamo = ingreso * 2

    else:
        prestamo = ingreso * 1

    print("\nRIESGO DEL CLIENTE:")
    print(riesgo[0])

    print("\nPRÉSTAMO RECOMENDADO:")
    print("$", prestamo)