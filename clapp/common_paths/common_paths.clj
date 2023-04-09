#!/usr/bin/env bb

(require '[clojure.string :as str]
         '[clojure.set :as set]
         '[babashka.fs :as fs])

(defn file-set [dir]
  (->> (fs/glob dir "**")
       (filter fs/regular-file?)
       (map (partial fs/relativize dir))
       (map str)
       set))

(defn common-files [dirs]
  (->> dirs
       (map file-set)
       (apply set/intersection)))

(defn validated-args [args]
    (let [argset (->> args (map fs/absolutize) (map fs/canonicalize) set)]
        (if (< (count argset) 2)
            (do (print "Error: Please provide at least two directories as arguments")
                (if (> (count args) (count argset))
                    (println " (some directories were duplicates)")
                    (println))
                (System/exit 1)))
        (let [invalid-dirs (->> argset (filter (complement fs/directory?)) (map str))]
        (if (> (count invalid-dirs) 0)
            (do (println "Error: invalid directories" invalid-dirs)
                (System/exit 1))))
        argset))

(let [dirs (validated-args *command-line-args*)]
    (doseq [f (common-files dirs)]
        (println f)))