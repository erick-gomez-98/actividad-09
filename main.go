package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type Alumno struct {
	Nombre       string
	Calificacion float64
	Materia      string
}

type Server struct {
	materias map[string]map[string]float64
	alumnos  map[string]map[string]float64
}

func (this *Server) AgregarCalificacion(al Alumno) error {
	if a, ok := this.alumnos[al.Nombre]; ok {
		// Ya existe, buscar la materia
		if _, o := a[al.Materia]; o {
			// Ya existe la materia, entonces no hay que permitir guardar
			return errors.New("Ya existe esa calificación para ese alumno y esa materia")
		}

		// No existe esa materia en ese alumno, agregarla
		this.alumnos[al.Nombre][al.Materia] = al.Calificacion
	} else {
		// No existe el alumno, generarlo
		materia := make(map[string]float64)
		materia[al.Materia] = al.Calificacion
		this.alumnos[al.Nombre] = materia
	}

	if _, ok := this.materias[al.Materia]; ok {
		this.materias[al.Materia][al.Nombre] = al.Calificacion
	} else {
		alumno := make(map[string]float64)
		alumno[al.Nombre] = al.Calificacion
		this.materias[al.Materia] = alumno
	}
	return nil
}

func (this *Server) PromedioAlumno(nombre string, reply *string) error {
	if al, ok := this.alumnos[nombre]; ok {
		// Si existe
		var sum float64
		for _, calificacion := range al {
			sum += calificacion
		}
		*reply = fmt.Sprint("El promedio de "+nombre+" es ", (sum / float64(len(al))))
		return nil
	}
	return errors.New("No existe ningun alumno con ese nombre")
}

func (this *Server) PromedioMateria(materia string, reply *string) error {
	if al, ok := this.materias[materia]; ok {
		// Si existe
		var sum float64
		for _, calificacion := range al {
			sum += calificacion
		}
		*reply = fmt.Sprint("El promedio de la materia de "+materia+" es ", (sum / float64(len(al))))
		return nil
	}
	return errors.New("No existe ninguna materia con ese nombre")
}

func (this *Server) PromedioGeneral(_ string, reply *string) error {
	var sum float64
	for _, materias := range this.alumnos {
		var promedioAlumno float64
		for _, calificacion := range materias {
			promedioAlumno += calificacion
		}
		promedioAlumno /= float64(len(materias))
		sum += promedioAlumno
	}
	*reply = fmt.Sprint("El promedio general es ", (sum / float64(len(this.alumnos))))
	return nil
}

var abc = ""

type Option struct {
	Name string
}

type AlumnosOptions struct {
	Alumnos []Option
}

type MateriasOptions struct {
	Materias []Option
}

type MateriaOption struct {
	Name         string
	Calificacion float64
}

type AlumnoPromedio struct {
	Msg     string
	Alumnos []Option
}

type MateriaPromedio struct {
	Msg      string
	Materias []Option
}

func promedioAlumno(res http.ResponseWriter, req *http.Request, server *Server) {
	tmpl := template.Must(template.ParseFiles("templates/promedio-alumno.html"))
	switch req.Method {
	case "GET":
		data := AlumnosOptions{}

		for alumno := range server.alumnos {
			data.Alumnos = append(data.Alumnos, Option{Name: alumno})
		}
		tmpl.Execute(res, data)

	case "POST":
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(res, "ParseForm() error %v", err)
			return
		}
		data := AlumnoPromedio{}
		for alumno := range server.alumnos {
			data.Alumnos = append(data.Alumnos, Option{Name: alumno})
		}
		var result string
		err := server.PromedioAlumno(req.FormValue("alumno"), &result)
		if err != nil {
			data := struct {
				Msg string
			}{
				Msg: "No se encontró el alumno",
			}
			tmpl.Execute(res, data)
		} else {
			data.Msg = result
			tmpl.Execute(res, data)
		}

	}
}

func promedioMateria(res http.ResponseWriter, req *http.Request, server *Server) {
	tmpl := template.Must(template.ParseFiles("templates/promedio-materia.html"))
	switch req.Method {
	case "GET":
		data := MateriasOptions{}

		for materia := range server.materias {
			data.Materias = append(data.Materias, Option{Name: materia})
		}
		tmpl.Execute(res, data)

	case "POST":
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(res, "ParseForm() error %v", err)
			return
		}
		data := MateriaPromedio{}
		for materia := range server.materias {
			data.Materias = append(data.Materias, Option{Name: materia})
		}

		var result string
		err := server.PromedioMateria(req.FormValue("materia"), &result)
		if err != nil {
			data := struct {
				Msg string
			}{
				Msg: "No se encontró la materia",
			}
			tmpl.Execute(res, data)
		} else {
			data.Msg = result
			tmpl.Execute(res, data)
		}
	}
}

func promedioGeneral(res http.ResponseWriter, req *http.Request, server *Server) {
	tmpl := template.Must(template.ParseFiles("templates/promedio-general.html"))
	switch req.Method {
	case "GET":
		var result string
		err := server.PromedioGeneral("", &result)
		if err != nil {
		data := struct {
			Msg string
		}{
			Msg: "Error al mostrar promedio general",
		}
		tmpl.Execute(res, data)
		} else {
			data := struct {
				Msg string
			}{
				Msg: result,
			}
			tmpl.Execute(res, data)
		}
	}
}

func agregar(res http.ResponseWriter, req *http.Request, server *Server) {
	tmpl := template.Must(template.ParseFiles("templates/agregar.html"))
	switch req.Method {
	case "GET":
		tmpl.Execute(res, nil)
	case "POST":
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(res, "ParseForm() error %v", err)
			return
		}

		al := Alumno{}
		al.Nombre = req.FormValue("nombre")
		i, e := strconv.ParseFloat(req.FormValue("calificacion"), 64)
		if e != nil {
			al.Calificacion = 0
		} else {
			al.Calificacion = i
		}
		al.Materia = req.FormValue("materia")

		err := server.AgregarCalificacion(al)
		if err != nil {
			data := struct {
				Msg string
			}{
				Msg: "Ya existe esa calificación para ese alumno y esa materia",
			}
			tmpl.Execute(res, data)
		} else {
			data := struct {
				Msg string
			}{
				Msg: "Calificación agregada",
			}
			tmpl.Execute(res, data)
		}

	}
}

func main() {
	s := new(Server)
	s.materias = make(map[string]map[string]float64)
	s.alumnos = make(map[string]map[string]float64)
	http.HandleFunc("/agregar", func(w http.ResponseWriter, r *http.Request) {
		agregar(w, r, s)
	})
	http.HandleFunc("/promedio-alumno", func(w http.ResponseWriter, r *http.Request) {
		promedioAlumno(w, r, s)
	})
	http.HandleFunc("/promedio-materia", func(w http.ResponseWriter, r *http.Request) {
		promedioMateria(w, r, s)
	})
	http.HandleFunc("/promedio-general", func(w http.ResponseWriter, r *http.Request) {
		promedioGeneral(w, r, s)
	})
	fmt.Println("Arrancando el servidor...")
	http.ListenAndServe(":9000", nil)
}
