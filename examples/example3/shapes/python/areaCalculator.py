from abc import ABC, abstractmethod
import math
import json

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

class AreaCalculator:
    def __init__(self, shapes=[]) -> None:
        self.shapes = shapes
    
    def sum(self):
        total_area = 0
        for shape in self.shapes:
            if isinstance(shape, Shape):
                total_area += shape.area()
            else:
                raise Exception("Invalid Shape: Shape does not implement Shape interface")
        return total_area
    
class VolumeCalculator(AreaCalculator):
    def sum(self):
        total_volume = 0
        for shape in self.shapes:
            total_volume += shape.area()
        return total_volume


class Square(Shape):
    def __init__(self, length) -> None:
        self.length = length
    
    def area(self):
        return self.length ** 2

class Circle(Shape):
    def __init__(self, radius) -> None:
        self.radius = radius

    def area(self):
        return math.pi * self.radius ** 2

class OutputFormatter:
    def __init__(self, areaCalculator: AreaCalculator) -> None:
        self.calculator = areaCalculator

    def to_json(self):
        data = {
            'sum': self.calculator.sum()
        }
        return json.dumps(data)

    def to_html(self):
        return f"Sum of the areas of provided shapes: {self.calculator.sum()}"


shapes = [Circle(2), Square(5), Square(6)]

# Area Calculation
area_calculator = AreaCalculator(shapes)
area_outputter = OutputFormatter(area_calculator)

print(area_outputter.to_json())
print(area_outputter.to_html())

# Volume Calculation (example)
solid_shapes = [Circle(3), Square(4)]
volume_calculator = VolumeCalculator(solid_shapes)
volume_outputter = OutputFormatter(volume_calculator)

print(volume_outputter.to_json())
print(volume_outputter.to_html())