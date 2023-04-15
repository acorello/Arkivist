#!/usr/bin/env bb

(require '[clojure.data.xml :as xml]
         '[clojure.java.io :as io])

(defn parse-ncx [file]
  (-> file
      io/reader
      xml/parse))

(defn navpoints [ncx]
  (->> ncx
       (map first)
    ;;    (filter #(= (:tag %) :navMap))
    ;;    first
    ;;    :content
    ;;    (filter #(= (:tag %) :navPoint))
       ))

(defn navpoint->markdown [navpoint]
  (->> navpoint
       :content
       (filter #(= (:tag %) :navLabel))
       first
       :content
       first
       :content
       (format "## %s")))

(defn ncx->markdown [ncx-file]
  (let [ncx (parse-ncx ncx-file)
        points (navpoints ncx)]
    (println points)
    #_(->> points
         (map navpoint->markdown)
         (str/join "\n"))))

;; Example usage:
(def toc-ncx "./OEBPS/toc.ncx")
(println (ncx->markdown toc-ncx))
