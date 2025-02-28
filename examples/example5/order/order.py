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


class SMSAuthorizer:
    def __init__(self):
        self.authenticated = False

    def verify_code(self, code: str):
        print("Verifying code")
        self.authenticated = True

    def is_authenticated(self):
        return self.authenticated


class PaymentProcessor(ABC):
    @abstractmethod
    def pay(self, order: Order):
        ...


class DebitPaymentProcessor(PaymentProcessor):
    def __init__(self, security_code: str, authorizer: SMSAuthorizer):
        self.security_code = security_code
        self.authorizer = authorizer

    def pay(self, order):
        if not self.authorizer.is_authenticated():
            raise Exception("Not authenticated")
        print("Processing debit payment type")
        print(f"Verifying security code: {self.security_code}")
        order.status = "paid"


class CreditPaymentProcessor(PaymentProcessor):
    def __init__(self, security_code: str):
        self.security_code = security_code

    def pay(self, order: Order):
        print("Processing credit payment type")
        print(f"Verifying security code: {self.security_code}")
        order.status = "paid"

class PaypalPaymentProcessor(PaymentProcessor):
    def __init__(self, security_code: str):
        self.security_code = security_code

    def pay(self, order, security_code: str):
        print("Processing paypal payment")
        print(f"Verifying email: {self.security_code}")
        order.status = "paid"


order = Order()

order.add_item("Keyboard", 1, 50)
order.add_item("SSD", 1, 150)
order.add_item("USB cable", 2, 5)

authorizer = SMSAuthorizer()
authorizer.verify_code("12345")

processor = DebitPaymentProcessor("0372846", authorizer)

processor.pay(order)

print(order.status)