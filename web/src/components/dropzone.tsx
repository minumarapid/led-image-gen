import {createSignal} from "solid-js";

export const [file, setFile] = createSignal<File | null>();

export const Dropzone = (props: {class: string}) => {
  const [isDragging, setIsDragging] = createSignal(false);
  const handleDragOver = (e: DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }

  const handleDragLeave = (e: DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    if (e.dataTransfer && e.dataTransfer.files.length > 0) {
      const droppedFile = e.dataTransfer.files[0];
      setFile(droppedFile);
      console.log("Dropped file:", droppedFile.name);
    }
  }
  return (
    <label
      class={`flex flex-col items-center justify-center h-full w-full transition-colors duration-300 border rounded-2xl border-mist-500 ${
        isDragging() ? "bg-[#64a296]/20" : "bg-transparent"
      } ${props.class && props.class}`}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <div class={`pointer-events-none flex flex-col gap-2 items-center justify-center`}>
        <span class={`text-center text-xl `}>画像をドラッグ＆ドロップ<br/>またはクリックして選択</span>
        <span class={`text-neutral-400 text-sm`}>対応形式：png, jpeg, gif, webp</span>
      </div>
      <input
        type="file"
        accept="image/png, image/jpeg, image/gif, image/webp"
        class="hidden"
        onChange={(e) => {
          const selectedFile = e.currentTarget.files?.[0];
          if (selectedFile) {
            setFile(selectedFile);
          }
          e.currentTarget.value = ""; // Reset the input
        }}
      />
    </label>
  )
}
