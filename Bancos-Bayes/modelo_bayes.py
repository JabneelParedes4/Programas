from sklearn.naive_bayes import GaussianNB
from sklearn.model_selection import cross_val_score

def entrenar_modelo(datos):

    X = datos[[
        "ingreso",
        "deuda",
        "score",
        "edad"
    ]]

    y = datos["riesgo"]

    # Modelo Bayesiano
    modelo = GaussianNB()

    # Cross Validation
    scores = cross_val_score(
        modelo,
        X,
        y,
        cv=5
    )

    print("\nCROSS VALIDATION")

    print("Scores:", scores)

    print("Accuracy promedio:", round(scores.mean(), 4))

    # Entrenar modelo final
    modelo.fit(X, y)

    return modelo