/**
 * getBase64FromBlob return base64 string from given blob
 * @param {object} blob
 */
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

/**
 * getFilenameFromResponse return the filename from the Content-Disposition in resonse
 * @param {object} response
 */
function getFilenameFromResponse(response) {
  const header = response.headers.get('Content-Disposition');
  const parts = header?.split(';');
  filename = parts[1].split('=')[1];
  return filename
}

/**
 * checkIsImageResponse check Content-Type in response whether contains "image" or not
 * @param {object} response
 */
function checkIsImageResponse(response) {
  const contenType = response.headers.get('Content-Type');
  return contenType?.includes("image") || false
}

/**
 * openUrlInNewWindow open the given url in new window
 * @param {string} url
 */
async function openUrlInNewWindow(url) {
  const response = await fetch(url)
  const blob = await response.blob()
  if (!checkIsImageResponse(response)) {
    return window.open(url)
  }
  const base64 = await getBase64FromBlob(blob)
  const image_window = window.open(url, "_blank")
  image_window.document.write(`<html><head></head><body><img width="400" src=${base64} alt="loading" /></body></html>`);
}

/**
 * addPreviewEle add the <img> for preview from given url
 * @param {string} url
 * @param {object} ele
 */
async function addPreviewEle(url, ele) {
  const response = await fetch(url)
  const blob = await response.blob()
  const filename = await getFilenameFromResponse(response)
  if (!checkIsImageResponse(response)) {
    const divTag = document.createElement("div");
      divTag.classList.add("text-[4rem]")
      divTag.classList.add("w-full")
      divTag.classList.add("text-center")
      divTag.classList.add("bx")

    if (filename.includes(".pdf")) {
      divTag.classList.add("bxs-file-pdf")
    } else {
      divTag.classList.add("bxs-file")
    }

    ele.appendChild(divTag)
    return
  }

  const base64 = await getBase64FromBlob(blob)
  const imgTag = document.createElement("img");
  imgTag.classList.add("w-full")
  imgTag.classList.add("h-full")
  imgTag.classList.add("object-cover")
  imgTag.src = base64
  ele.appendChild(imgTag)
}

/**
 * fileToBase64 convert the selected File object to base64 (encode data to Base64)
 * @param {File} file
 * @return {Promise}
 */
const fileToBase64 = async file => new Promise((resolve, reject) => {
  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.onload = () => resolve(reader.result);
  reader.onerror = reject;
});
