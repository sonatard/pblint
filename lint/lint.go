package lint

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/iancoleman/strcase"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func Lint(files []*desc.FileDescriptor) []error {
	var errs []error

	for _, file := range files {
		for _, service := range file.GetServices() {
			for _, method := range service.GetMethods() {
				// Lint rules
				if err := fileName(file, service); err != nil {
					errs = append(errs, err)
				}
				if err := requestTypeName(method); err != nil {
					errs = append(errs, err)
				}
				if err := responseTypeName(method); err != nil {
					errs = append(errs, err)
				}
				if err := requestTypeInFile(file, method); err != nil {
					errs = append(errs, err)
				}
				if err := responseTypeInFile(file, method); err != nil {
					errs = append(errs, err)
				}

				httpRule, err := httpRule(method)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if err := httpMethod(httpRule); err != nil {
					errs = append(errs, err)
				}
				if err := httpBody(httpRule); err != nil {
					errs = append(errs, err)
				}
				if err := httpAdditionalBinding(httpRule); err != nil {
					errs = append(errs, err)
				}
				if err := httpURL(service, method, httpRule); err != nil {
					errs = append(errs, err)
				}
			}
		}

		for _, messageType := range file.GetMessageTypes() {
			if err := modelMessageInFile(messageType); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func fileName(file *desc.FileDescriptor, service *desc.ServiceDescriptor) error {
	wantFileName := strcase.ToSnake(service.GetName()) + ".proto"
	paths := strings.Split(file.GetName(), "/")
	gotFileName := paths[len(paths)-1]

	if wantFileName != gotFileName {
		return fmt.Errorf("error: file name must be %v, got=%v", wantFileName, gotFileName)
	}

	return nil
}

func requestTypeInFile(file *desc.FileDescriptor, method *desc.MethodDescriptor) error {
	wantFileName := file.GetName()
	gotFileName := method.GetInputType().GetFile().GetName()
	if wantFileName != gotFileName {
		return fmt.Errorf("error: request type must be in %v, got=%v", wantFileName, gotFileName)
	}

	return nil
}

func responseTypeInFile(file *desc.FileDescriptor, method *desc.MethodDescriptor) error {
	wantFileName := file.GetName()
	gotFileName := method.GetOutputType().GetFile().GetName()
	if wantFileName != gotFileName {
		return fmt.Errorf("error: response type must be in %v, got=%v", wantFileName, gotFileName)
	}

	return nil
}

func httpRule(method *desc.MethodDescriptor) (*annotations.HttpRule, error) {
	opts := method.GetOptions()

	if !proto.HasExtension(opts, annotations.E_Http) {
		return nil, fmt.Errorf("error: HTTP Rule not found")
	}

	ext, err := proto.GetExtension(opts, annotations.E_Http)
	if err != nil {
		return nil, fmt.Errorf("error: HTTP Rule not found")
	}

	rule, ok := ext.(*annotations.HttpRule)
	if !ok {
		return nil, fmt.Errorf("error: HTTP Rule not found")
	}

	return rule, nil
}

func httpMethod(httpRule *annotations.HttpRule) error {
	switch httpRule.GetPattern().(type) {
	case *annotations.HttpRule_Get, *annotations.HttpRule_Post:
		return nil
	}

	return fmt.Errorf("error: HTTP Rule HTTP method must use GET or POST")
}

func httpURL(service *desc.ServiceDescriptor, method *desc.MethodDescriptor, httpRule *annotations.HttpRule) error {
	wantURL := fmt.Sprintf("/%v/%v", service.GetName(), method.GetName())
	switch opt := httpRule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		gotURL := opt.Get
		if gotURL != wantURL {
			return fmt.Errorf("error: HTTP want=%v, got=%v", wantURL, gotURL)
		}
	case *annotations.HttpRule_Post:
		gotURL := opt.Post
		if gotURL != wantURL {
			return fmt.Errorf("error: HTTP want=%v, got=%v", wantURL, gotURL)
		}
	}

	return nil
}

func httpBody(httpRule *annotations.HttpRule) error {
	body := httpRule.GetBody()
	if body != "*" {
		return fmt.Errorf("error: HTTP Rule Body is not *. got=%v", body)
	}

	return nil
}

func httpAdditionalBinding(httpRule *annotations.HttpRule) error {
	if len(httpRule.GetAdditionalBindings()) != 0 {
		return fmt.Errorf("error: HTTP Rule must not use AdditionalBindingsis")
	}

	return nil
}

func requestTypeName(method *desc.MethodDescriptor) error {
	methodName := method.GetName()

	wantRequestName := methodName + "Request"
	gotRequestName := method.GetInputType().GetName()
	if gotRequestName != wantRequestName {
		return fmt.Errorf("error: RequestName want=%v, got=%v", wantRequestName, gotRequestName)
	}

	return nil
}

func responseTypeName(method *desc.MethodDescriptor) error {
	methodName := method.GetName()

	wantResponseName := methodName + "Response"
	gotResponseName := method.GetOutputType().GetName()
	if gotResponseName != wantResponseName {
		return fmt.Errorf("error: ResponseName want=%v, got=%v", wantResponseName, gotResponseName)
	}

	return nil
}

func modelMessageInFile(message *desc.MessageDescriptor) error {
	messageName := message.GetName()
	if strings.HasSuffix(messageName, "Request") {
		return nil
	}
	if strings.HasSuffix(messageName, "Response") {
		return nil
	}

	gotFileName := message.GetFile().GetName()
	if strings.HasSuffix(gotFileName, "_service.proto") {
		return fmt.Errorf("error: model message must not be in _service.go, got=%v", gotFileName)
	}

	return nil
}
