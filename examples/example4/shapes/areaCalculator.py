from abc import ABC, abstractmethod
import math
import json


class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

    @abstractmethod
    def volume(self):
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
            total_volume += shape.volume()
        return total_volume

class Square(Shape):
    def __init__(self, length) -> None:
        self.length = length
    
    def area(self):
        return self.length ** 2
    
    def volume(self):
        raise NotImplementedError("The volume cannot be calculated.")

class Circle(Shape):
    def __init__(self, radius) -> None:
        self.radius = radius

    def area(self):
        return math.pi * self.radius ** 2
    
    def volume(self):
        raise NotImplementedError("The volume cannot be calculated")
    
class Dice(Shape):
    def __init__(self, length) -> None:
        self.length = length
    
    def area(self):
        return 6 ** pow(self.length, 2) 
    
    def volume(self):
        return 6 ** pow(self.length, 3)
    
class SumCalculator:
    def __init__(self, areaCalculator: AreaCalculator) -> None:
        self.calculator = areaCalculator

    def JSON(self):
        data = {
            'sum': self.calculator.sum()
        }
        return json.dumps(data)

    def HTML(self):
        return f"Sum of the areas of provided shapes: {self.calculator.sum()}"
    
shapes = [Circle(2), Square(5), Square(6), Dice(2)]

# Area Calculation
area_calculator = AreaCalculator(shapes)
area_outputter = SumCalculator(area_calculator)

print(area_outputter.to_json())
print(area_outputter.to_html())

# Volume Calculation (example)
volume_calculator = VolumeCalculator(shapes)
volume_outputter = SumCalculator(volume_calculator)

print(volume_outputter.to_json())
print(volume_outputter.to_html())