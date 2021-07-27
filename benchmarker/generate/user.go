package generate

import (
	"bufio"
	"os"
	"strings"

	"github.com/isucon/isucon11-final/benchmarker/model"
)

var (
	studentFile = "./generate/data/student.tsv"
	facultyFile = "./generate/data/faculty.tsv"
)

func LoadStudentsData() ([]*model.UserAccount, error) {
	return loadUserAccountData(studentFile)
}

func LoadFacultiesData() ([]*model.UserAccount, error) {
	return loadUserAccountData(facultyFile)
}

func loadUserAccountData(path string) ([]*model.UserAccount, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	userDataSet := make([]*model.UserAccount, 0)
	s := bufio.NewScanner(file)
	for i := 0; s.Scan(); i++ {
		line := strings.Split(s.Text(), "\t")
		code := line[0]
		name := line[1]
		rawPW := line[2]

		account := &model.UserAccount{
			Code:        code,
			Name:        name,
			RawPassword: rawPW,
		}
		userDataSet = append(userDataSet, account)
	}

	return userDataSet, nil
}
