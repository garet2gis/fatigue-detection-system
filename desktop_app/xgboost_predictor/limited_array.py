class LimitedSizeArray:
    def __init__(self, size):
        self.size = size
        self.array = [0] * size
        self.zero_count = size  # Изначально все элементы равны 0

    def push(self, value):
        # Вытеснение старого элемента
        if len(self.array) == self.size:
            self.zero_count -= int(self.array.pop(0) == 0)

        # Добавление нового элемента
        self.array.append(value)

        # Обновление счетчика нулей
        self.zero_count += int(value == 0)

    def count_zeros(self):
        return self.zero_count