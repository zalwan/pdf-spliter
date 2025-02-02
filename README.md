# PDF Splitter Web App

## Description

A web application built with **Golang** and the **Gin** framework to split PDF files into separate pages. Each page of the PDF is extracted as an individual file and can be renamed based on user input. The split files are compressed into a ZIP archive for easy download.

## Features

- Upload a PDF file and split each page into a separate file.
- Option to assign custom names to the extracted pages.
- Generate a ZIP file containing all split pages.

## Technologies Used

- **Golang**
- **Gin (Web Framework)**
- **pdfcpu (PDF manipulation library)**
- **HTML**

## Installation and Usage

### 1. Clone the Repository

```sh
git clone https://github.com/zalwan/pdf-splitter.git
cd pdf-splitter
```

### 2. Install Dependencies

Ensure you have **Golang** installed.

```sh
go mod tidy
```

### 3. Run the Application

```sh
go run main.go
```

The application will be available at `http://localhost:8080/`

### 4. Using the Application

1. Open your browser and go to `http://localhost:8080/`.
2. Upload a PDF file.
3. Enter a list of names for the split pages.
4. Click the **Split PDF** button.
5. Download the ZIP file containing the split pages.

## Directory Structure

```
/
|-- uploads/         # Folder to store uploaded files
|-- temp_outputs/    # Temporary folder for split files
|-- templates/       # HTML templates for the web interface
|-- main.go          # Main application file
|-- go.mod           # Dependency management file
|-- README.md        # Project documentation
```

## Contributors

- **Rizal Suryawan** (Feel free to edit this with your name)

## License

This project is licensed under the **MIT License**. You are free to use it with proper attribution.

---

**Happy Coding! 🚀**
