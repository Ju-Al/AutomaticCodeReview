class ValidatePerson:
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def validate_name(self, name):
        if len(name) > 3:
            return True
        else:
            return False

    def validate_age(self, age):
        if age > 18:
            return True
        else:
            return False

    def display(self):
        if self.validate_name(self.name) and self.validate_age(self.age):
            print(f"Name: {self.name} and Age: {self.age}")
        else:
            print('Invalid')


# Beispielaufruf
person = ValidatePerson("John", 20)
person.display()