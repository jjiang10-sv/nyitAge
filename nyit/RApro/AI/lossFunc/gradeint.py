import numpy as np
import matplotlib.pyplot as plt

# Define the loss function L(w) = (w - 3)^2
def loss(w):
    return (w - 3) ** 2

# Derivative of the loss function
def gradient(w):
    return 2 * (w - 3)

# Gradient descent parameters
learning_rate = 0.1
iterations = 20
w = 10  # initial weight
w_values = [w]
loss_values = [loss(w)]

# Perform gradient descent
for _ in range(iterations):
    grad = gradient(w)
    w = w - learning_rate * grad
    w_values.append(w)
    loss_values.append(loss(w))

# Plotting the loss function and the descent steps
w_range = np.linspace(0, 12, 100)
loss_range = loss(w_range)

plt.figure(figsize=(10, 6))
plt.plot(w_range, loss_range, label='Loss Function: L(w) = (w - 3)^2')
plt.scatter(w_values, loss_values, color='red', label='Descent Steps')
plt.plot(w_values, loss_values, color='red', linestyle='--')
plt.title('Gradual Gradient Descent on Loss Function')
plt.xlabel('w (parameter)')
plt.ylabel('Loss')
plt.legend()
plt.grid(True)
plt.show()
