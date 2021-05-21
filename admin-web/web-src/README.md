# preact.js starter project

## 3步驟使用

1. 安裝相依套件: `npm install`
2. 啟動開發模式: `npm run start`
3. 瀏覽器打開: [http://127.0.0.1:8000/](http://127.0.0.1:8000/)

## 檔案架構
	* `src/` : code
		* `main.js` : entry
		* `comp.jsx` : 組件, jsx版
		* `comp.js` : 組件, js版
	* `public/` : html & 圖片等共用資源

## 基本用法

* 開發
	* esbuild:
		* 啟動&重編反應快速
		* `npm run start`
		* 瀏覽器打開: [http://127.0.0.1:8008/](http://127.0.0.1:8008/)
		* 無限F5
* 打包輸出
	* `npm run build`
	* 輸出在 `dist/` 資料夾內

### dependencies

* esbuild:
	* 已內建jsx轉譯&打包
	* 缺點不可轉譯到ES5以下的js語法
	* 自動把所有相關資源複製制輸出資料夾可能要自己另外處理 (只打包js)
	* 可以獨立使用(不與webpack組合)

* 其他
	* `cross-env` : 方便切換設定用
