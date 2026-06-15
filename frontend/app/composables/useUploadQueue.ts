export interface UploadTask {
  id: string;
  fileName: string;
  loaded: number;
  total: number;
  percentage: number;
  speed: number;
  remainingSeconds: number;
  status: "pending" | "uploading" | "done" | "error" | "cancelled";
  xhr?: XMLHttpRequest;
}

export function useUploadQueue() {
  const uploads = useState<UploadTask[]>("upload-queue", () => []);

  function upload(
    body: FormData | File | Blob,
    url: string,
    options: {
      method?: string;
      headers?: Record<string, string>;
      withCredentials?: boolean;
    } = {},
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      const id = Math.random().toString(36).slice(2);
      let fileName = "file";
      if (body instanceof File) {
        fileName = body.name;
      } else if (body instanceof FormData) {
        const f = body.get("file");
        if (f instanceof File) fileName = f.name;
      }

      const task: UploadTask = {
        id,
        fileName,
        loaded: 0,
        total: 0,
        percentage: 0,
        speed: 0,
        remainingSeconds: 0,
        status: "pending",
      };

      uploads.value.push(task);

      const xhr = new XMLHttpRequest();
      task.xhr = xhr;

      const startTime = Date.now();
      let lastLoaded = 0;
      let lastTime = startTime;

      xhr.upload.addEventListener("progress", (e) => {
        if (!e.lengthComputable) return;
        const now = Date.now();
        const dt = (now - lastTime) / 1000;
        if (dt <= 0) return;

        const dloaded = e.loaded - lastLoaded;
        task.loaded = e.loaded;
        task.total = e.total;
        task.percentage = Math.round((e.loaded / e.total) * 100);
        task.status = "uploading";
        task.speed = dloaded / dt;
        task.remainingSeconds = Math.max(0, (e.total - e.loaded) / task.speed);

        lastLoaded = e.loaded;
        lastTime = now;
      });

      xhr.addEventListener("load", () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          task.status = "done";
          task.percentage = 100;
          setTimeout(() => {
            uploads.value = uploads.value.filter((u) => u.id !== id);
          }, 2000);
          resolve();
        } else {
          task.status = "error";
          uploads.value = uploads.value.filter((u) => u.id !== id);
          reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
        }
      });

      xhr.addEventListener("error", () => {
        task.status = "error";
        uploads.value = uploads.value.filter((u) => u.id !== id);
        reject(new Error("Network error during upload"));
      });

      xhr.addEventListener("abort", () => {
        task.status = "cancelled";
        uploads.value = uploads.value.filter((u) => u.id !== id);
        reject(new Error("Upload cancelled"));
      });

      xhr.open(options.method || "POST", url);
      xhr.withCredentials = options.withCredentials ?? false;
      if (options.headers) {
        Object.entries(options.headers).forEach(([k, v]) =>
          xhr.setRequestHeader(k, v),
        );
      }

      xhr.send(body);
    });
  }

  function cancelUpload(id: string) {
    const task = uploads.value.find((u) => u.id === id);
    if (task?.xhr) {
      task.xhr.abort();
    }
  }

  return { uploads, upload, cancelUpload };
}
