package main

import (
	"database/sql"
	f "fmt"
    "bufio"
	"os"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"github.com/joho/godotenv"


)

type UserSession struct {
		UserId  uint;
		Username string;
		HashPassword string;
}

type Task struct {
	task_id uint;
	username uint;
	task     string;
	status   uint8;
	date     string

}

func main() {
	f.Println("welcome to taskmanagerV2");
	// remote mysql server connection 
	err := godotenv.Load()
    if err != nil {
        f.Println("error loading .env file");
    }
	dns := os.Getenv("dns");
	db, err := sql.Open("mysql", dns);
	if err != nil {
		panic(err.Error());
	}
	defer db.Close();
	for {
		f.Println("to register press r || to login press l || to exit press e");
		var username string;
		var password string;
		var input string;
		f.Scanln(&input);
		tableU := os.Getenv("tableU");
		tableT := os.Getenv("tableT");
		switch input {
			case "r":
				f.Println("enter your username:");
				f.Scanln(&username);
				f.Println("enter your password");
				f.Scanln(&password);
				hashedPassword, err := hashPassword(password);
				if err != nil {
					f.Println("error during hashing");
				}
				stmt, err := db.Prepare("INSERT INTO " + tableU+" (username, password) VALUES (?, ?)");
					if err != nil {
						f.Println("error during preparing", err);
					}
			
				_, err = stmt.Exec(username, hashedPassword);
				
				if err != nil {
					f.Println("error", err);
				}
				stmt.Close();
				f.Println("sucesful registration");

			case "l":
				f.Println("enter your username");
				f.Scanln(&username);
				f.Println("enter your password");
				f.Scanln(&password);
				stmt, err := db.Prepare("SELECT * FROM "+ tableU +" WHERE username = ?");
				if err != nil {
					f.Println("error during prepare", err);
				}
				
		
				var UserSession UserSession;
			
				err = stmt.QueryRow(username).Scan(&UserSession.UserId, &UserSession.Username, &UserSession.HashPassword);
				if err != nil {
					if err == sql.ErrNoRows {
						f.Println("User not found")
					} else {
						f.Println("Error during query:", err)
					}
					return
				}
			
				if err != nil {
					f.Println("Error during hashing:", err)
					return
				}
				if !comparePasswords(password, UserSession.HashPassword) {
					f.Println("invalid password");
				} else {
					f.Println("Succesful login:", UserSession.Username)
					var loginInput string;
					f.Println("press v to view tasks || press t to add task  press u to update task || press d to delete task || press x to logout");
					InnerLoop:
					for {
						f.Scanln(&loginInput);
						switch loginInput {
						case "t" :
							f.Println("add task");
							var task string;
							scanner := bufio.NewScanner(os.Stdin);
							if scanner.Scan() {
								task = scanner.Text();
							} else {    
								f.Println("error reading input ", scanner.Err());
								return;
							}
							stmt, err := db.Prepare("INSERT INTO " + tableT +" (username, task) VALUES (?, ?)");
							if err != nil {
								f.Println("error during preparing", err);
							}
							_, err = stmt.Exec(UserSession.UserId, task);
							if err != nil {
								f.Println(err);
							}

							stmt.Close();


							f.Println("task added");
							f.Println("press v to view tasks || press t to add task  press u to update task || press d to delete task || press x to logout");
						case "u":
							f.Println("Select id of task to update:");
							viewTasks(UserSession.UserId, db)
							var selectedId uint;
							f.Scanln(&selectedId);
							f.Println("edit task description (press enter if dont want to change it)");
							var task string;
							var status string;
							var sql string;
							scanner := bufio.NewScanner(os.Stdin);
							if scanner.Scan() {
								task = scanner.Text();
							} else {    
								f.Println("error reading input ", scanner.Err());
					
								return;
							}
							f.Println("edit status (0 = not finished, 1 = finished)");
							f.Scanln(&status);
							args := [] interface{}{};
							if task == "" {
								sql = "UPDATE " + tableT +" SET status = ? WHERE task_id = ? AND username = ?";
								args = append(args, status , selectedId, UserSession.UserId);
								} else {
									sql = "UPDATE " + tableT+ " SET task = ?, status = ? WHERE task_Id = ? AND username = ?";
									args = append(args, task, status, selectedId, UserSession.UserId);
								
								}
								stmt, err := db.Prepare(sql);
								if err != nil {
									f.Println("err", err);
								}
								defer stmt.Close();

								results, err := stmt.Exec(args...);
								if err != nil {
									f.Println("err", err);
								}
								rowAffected, err := results.RowsAffected();
								if err != nil {
									f.Println("err", err);
								}
								if rowAffected == 0 {
									f.Println("incorrect input try again");
								} else {
									f.Printf("updated tasks with %d, rows affected %d \n", selectedId, rowAffected);

								}
								f.Println("press v to view tasks || press t to add task  press u to update task || press d to delete task || press x to logout");
					
						case "d":
							f.Println("Select id of task to update:");
							viewTasks(UserSession.UserId, db)
							var selectedId uint;
							f.Scanln(&selectedId);
							stmt, err := db.Prepare("DELETE FROM " + tableT +" WHERE task_id = ?");
							if err != nil {
								f.Println("err", err);
							}

							defer stmt.Close();
							results, err := stmt.Exec(selectedId);
							if err != nil {
								f.Println("err", err);
							}
							rowsAffected , err := results.RowsAffected();
							if err != nil {
								f.Println("err", err);
							}

							if rowsAffected > 0 {
								f.Println("deleted task with id ", selectedId);
							}

							

							f.Println("press v to view tasks || press t to add task  press u to update task || press d to delete task || press x to logout");
						case "v":
							viewTasks(UserSession.UserId, db);
							f.Println("\n press v to view tasks || press t to add task  press u to update task || press d to delete task || press x to logout");

						case "x":
							f.Println("logout");
							break InnerLoop;
						default:
							f.Println("select proper key");
						}
			
					
					}
			
				
				}

				stmt.Close();
			

			case "e":
				return;
			}


		}
	

}




func viewTasks(userId uint, db *sql.DB ) {
	f.Println("view tasks");
	tableT := os.Getenv("tableT");
	stmt, err := db.Prepare("SELECT * FROM " +tableT+" WHERE username = ?");
	 if err != nil {
		f.Println(err);
	 }
	 defer stmt.Close()

	 rows ,err := stmt.Query(userId);
	 if err != nil {
		f.Println(err);
	 }

	 defer rows.Close()

	 var results []Task;
	 
	 for rows.Next() {
		var result Task;
		err := rows.Scan(&result.task_id, &result.username, &result.task,&result.date, &result.status);
		if err != nil {
			f.Println(err);	
		}
		results = append(results, result);
	 }
			 
	 for _, r := range results {
		f.Printf("task id %d user id %d task descrpition %s: date %s status: %d \n", r.task_id, r.username, r.task, r.date, r.status);
	}

}

func hashPassword (password string) (string ,error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost);
	if err != nil {
		return "err", err;
	}
	return string(hash), nil;
}


func comparePasswords(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password));
	return err == nil;
}