import math

class AreaCalculator:

    def __init__(self, shapes=[]) -> None:
        self.shapes = shapes
    
    def sum(self):
        areas = []
        for shape in self.shapes:
            if isinstance(shape, Square):
                areas.append(math.pow(shape.length, 2))
            elif isinstance(shape, Circle):
                areas.append(math.pi * math.pow(shape.radius, 2))
        return sum(areas)
    
    def output(self):
        return f"Sum of the areas of provided shapes: {self.sum()}"

class Square:
    def __init__(self, length) -> None:
        self.length = length

class Circle:
    def __init__(self, radius) -> None:
        self.radius = radius


shapes = [
    Circle(2),
    Square(5),
    Square(6),
]

areas = AreaCalculator(shapes)
print(areas.output())
