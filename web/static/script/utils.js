async function getBase64FromBlob(blob) {
  const base64 = await new Promise((resolve) => {
    const reader = new FileReader();
    reader.readAsDataURL(blob);
    reader.onloadend = () => {
      const base64data = reader.result;
      resolve(base64data);
    }
  });
  return base64
}

async function getFilenameFromResponse(response) {
  const header = response.headers.get('Content-Disposition');
  const parts = header?.split(';');
  filename = parts[1].split('=')[1];
  return filename
}
