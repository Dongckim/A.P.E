export type MediaKind = "image" | "video" | "audio" | "pdf";

const KIND_BY_EXT: Record<string, MediaKind> = {
  png: "image", jpg: "image", jpeg: "image", gif: "image", svg: "image", webp: "image",
  mp4: "video", webm: "video",
  mp3: "audio", wav: "audio",
  pdf: "pdf",
};

export function getMediaKind(name: string): MediaKind | null {
  const ext = name.split(".").pop()?.toLowerCase() || "";
  return KIND_BY_EXT[ext] ?? null;
}
