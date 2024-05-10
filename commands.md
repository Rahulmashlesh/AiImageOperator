# Useful commands for this project

## Switch context to colima
```sh
kubectx colima
```

## Build

```sh 
 make generate 
 make docker-build
 make install
 make deploy
```

```sh
make docker-build
make install
make deploy
```

## Rollout
```sh 

 k rollout restart deployment/aiimageoperator-controller-manager -n aiimageoperator-system

```

## TODO:
implement inteface for saving. 
redis in diff name space. 