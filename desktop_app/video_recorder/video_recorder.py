from PyQt5.QtCore import QThread, pyqtSignal
from vidgear.gears import CamGear, WriteGear
import uuid
import time


class VideoRecorder(QThread):
    finished = pyqtSignal(str, str)  # Сигнал о завершении записи

    def __init__(self, video_len_sec=15):
        super().__init__()
        self.video_len_sec = video_len_sec

    def run(self):
        options = {
            "CAP_PROP_FRAME_WIDTH": 244,
            "CAP_PROP_FRAME_HEIGHT": 244,
            "CAP_PROP_FPS": 244,
        }

        stream = CamGear(source=1, **options).start()

        video_id = str(uuid.uuid4())
        filename = f"{video_id}.mp4"

        writer = WriteGear(output=filename)

        start = time.perf_counter()
        while True:
            frame = stream.read()

            if frame is None:
                break

            writer.write(frame)

            stop = time.perf_counter()
            if stop - start > self.video_len_sec:
                break

        stream.stop()
        writer.close()
        self.finished.emit(video_id, filename)
