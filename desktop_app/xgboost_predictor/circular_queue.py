class CircularQueue:
    def __init__(self, capacity: int):
        if capacity <= 0:
            raise ValueError("capacity должен быть больше нуля")
        self.capacity = capacity
        self.queue = [None] * capacity
        self.head = 0
        self.tail = 0
        self.size = 0

    def is_full(self) -> bool:
        """Проверяет, полна ли очередь."""
        return self.size == self.capacity

    def is_empty(self) -> bool:
        """Проверяет, пуста ли очередь."""
        return self.size == 0

    def enqueue(self, item):
        """Добавляет элемент в конец очереди, при переполнении удаляет элемент с начала."""
        if self.is_full():
            self.dequeue()  # Удаляем элемент из начала, если очередь переполнена
        self.queue[self.tail] = item
        self.tail = (self.tail + 1) % self.capacity
        if self.size < self.capacity:
            self.size += 1

    def dequeue(self):
        """Удаляет элемент из начала очереди и возвращает его."""
        if self.is_empty():
            raise IndexError("Очередь пуста")
        item = self.queue[self.head]
        self.queue[self.head] = None  # Очистка ссылки для GC, если это необходимо
        self.head = (self.head + 1) % self.capacity
        self.size -= 1
        return item

    def get_raw_array(self):
        """Возвращает массив очереди без None элементов."""
        return [item for item in self.queue if item is not None]
