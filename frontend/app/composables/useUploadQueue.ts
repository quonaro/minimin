export interface UploadTask {
  id: string;
  fileName: string;
  loaded: number;
  total: number;
  percentage: number;
  speed: number;
  remainingSeconds: number;
  duration?: number; // ms, set on done
  status: "pending" | "uploading" | "done" | "error" | "cancelled";
  xhr?: XMLHttpRequest;
}

export function useUploadQueue() {
  const uploads = useState<UploadTask[]>("upload-queue", () => []);

  function updateTask(id: string, patch: Partial<UploadTask>) {
    const t = uploads.value.find((u) => u.id === id);
    if (!t) return;
    Object.assign(t, patch);
  }

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

      const fileSize =
        body instanceof File
          ? body.size
          : body instanceof FormData
            ? (body.get("file") instanceof File
                ? (body.get("file") as File).size
                : 0)
            : 0;

      uploads.value.push({
        id,
        fileName,
        loaded: 0,
        total: fileSize,
        percentage: 0,
        speed: 0,
        remainingSeconds: 0,
        status: "pending",
      });

      const xhr = new XMLHttpRequest();
      updateTask(id, { xhr });

      const startTime = Date.now();
      let lastLoaded = 0;
      let lastTime = startTime;

      xhr.upload.addEventListener("progress", (e) => {
        const t = uploads.value.find((u) => u.id === id);
        if (!t) return;
        const total = e.total && e.total > 0 ? e.total : t.total;
        if (!total) return;

        const now = Date.now();
        const dt = (now - lastTime) / 1000;
        if (dt <= 0) return;

        const dloaded = e.loaded - lastLoaded;
        const speed = dloaded / dt;
        updateTask(id, {
          loaded: e.loaded,
          total,
          percentage: Math.min(100, Math.round((e.loaded / total) * 100)),
          status: "uploading",
          speed,
          remainingSeconds:
            speed > 0 ? Math.max(0, (total - e.loaded) / speed) : 0,
        });

        lastLoaded = e.loaded;
        lastTime = now;
      });

      xhr.addEventListener("load", () => {
        const duration = Date.now() - startTime;
        if (xhr.status >= 200 && xhr.status < 300) {
          updateTask(id, {
            status: "done",
            percentage: 100,
            duration,
          });
          resolve();
        } else {
          updateTask(id, { status: "error", duration });
          reject(
            new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`),
          );
        }
      });

      xhr.addEventListener("error", () => {
        updateTask(id, {
          status: "error",
          duration: Date.now() - startTime,
        });
        reject(new Error("Network error during upload"));
      });

      xhr.addEventListener("abort", () => {
        updateTask(id, {
          status: "cancelled",
          duration: Date.now() - startTime,
        });
        reject(new Error("Upload cancelled"));
      });

      xhr.open(options.method || "POST", url);
      xhr.withCredentials = options.withCredentials ?? false;
      if (options.headers) {
        Object.entries(options.headers).forEach(([k, v]) =>
          xhr.setRequestHeader(k, v),
        );
      }

      let sendBody: XMLHttpRequestBodyInit = body;
      if (body instanceof File || body instanceof Blob) {
        const fd = new FormData();
        if (body instanceof File) {
          fd.append("file", body, body.name);
        } else {
          fd.append("file", body);
        }
        sendBody = fd;
      }

      xhr.send(sendBody);
    });
  }

  function cancelUpload(id: string) {
    const task = uploads.value.find((u) => u.id === id);
    if (task?.xhr) {
      task.xhr.abort();
    }
  }

  function removeUpload(id: string) {
    uploads.value = uploads.value.filter((u) => u.id !== id);
  }

  return { uploads, upload, cancelUpload, removeUpload };
}
