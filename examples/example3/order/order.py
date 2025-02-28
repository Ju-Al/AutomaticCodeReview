
from abc import ABC, abstractmethod

class Order:
    def __init__(self):
        self.items = []
        self.quantities = []
        self.prices = []
        self.status = "open"

    def add_item(self, name: str, quantity: int, price: float) -> None:
        self.items.append(name)
        self.quantities.append(quantity)
        self.prices.append(price)


class PaymentProcessor(ABC):
    @abstractmethod
    def pay(self, order, security_code: str):
        pass


class CreditCardPaymentProcessor(PaymentProcessor):
    def pay(self, order, security_code: str):
        print("Processing credit card payment")
        print(f"Verifying security code: {security_code}")
        order.status = "paid"


class DebitCardPaymentProcessor(PaymentProcessor):
    def pay(self, order, security_code: str):
        print("Processing debit card payment")
        print(f"Verifying security code: {security_code}")
        order.status = "paid"


class PaypalPaymentProcessor(PaymentProcessor):
    def pay(self, order, security_code: str):
        print("Processing paypal payment")
        print(f"Verifying email: {security_code}")
        order.status = "paid"


order = Order()

order.add_item("Keyboard", 1, 50)
order.add_item("SSD", 1, 150)
order.add_item("USB cable", 2, 5)

processor = CreditCardPaymentProcessor()
processor.pay(order, "0372846")
print(order.status)
